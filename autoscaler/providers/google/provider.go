package google

import (
	"context"
	"fmt"
	"github.com/laszlocph/woodpecker/autoscaler"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/cloudresourcemanager/v1"
	compute "google.golang.org/api/compute/v1"
	"strconv"
	"strings"
	"time"
)

type provider struct {
	project       string
	projectId     string
	zone          string
	instanceGroup string
	service       *compute.Service
}

type Option func(*provider)

func WithProject(project string) Option {
	return func(p *provider) {
		p.project = project
	}
}

func WithZone(zone string) Option {
	return func(p *provider) {
		p.zone = zone
	}
}

func WithInstanceGroup(instanceGroup string) Option {
	return func(p *provider) {
		p.instanceGroup = instanceGroup
	}
}

func New(opts ...Option) (autoscaler.Provider, error) {
	p := new(provider)
	for _, opt := range opts {
		opt(p)
	}
	client, err := google.DefaultClient(context.Background(), compute.ComputeScope)
	if err != nil {
		logrus.WithError(err).Warn("Failed to create gcp oauth2 client")
		return nil, err
	}

	resourceManager, err := cloudresourcemanager.New(client)
	if err != nil {
		logrus.WithError(err).Warn("Failed to create resource manager client")
		return nil, err
	}

	project, err := resourceManager.Projects.Get(p.project).Do()
	if err != nil {
		logrus.WithError(err).Warn("Failed to get project")
		return nil, err
	}
	p.projectId = strconv.FormatInt(project.ProjectNumber, 10)

	p.service, err = compute.New(client)
	if err != nil {
		logrus.WithError(err).Warn("Failed to create compute client")
		return nil, err
	}

	return p, nil
}

func (p provider) SetCapacity(n int, minimumAge time.Duration) error {
	instanceGroup, err := p.service.InstanceGroupManagers.Get(p.project, p.zone, p.instanceGroup).Do()
	if err != nil {
		logrus.WithError(err).Error("Failed to get instance group")
		return err
	}
	currentSize := int(instanceGroup.TargetSize)
	if currentSize == n {
		logrus.Debug("Not autoscaling as desired size is target size")
		return nil
	}

	if n < currentSize {
		filter := fmt.Sprintf("projects/%s/zones/%s/instanceGroupManagers/%s", p.projectId, p.zone, p.instanceGroup)
		var workers []*compute.Instance
		instanceList, err := p.service.Instances.List(p.project, p.zone).
			Do()
		if err != nil {
			logrus.WithError(err).Error("Failed to list instances")
			return err
		}
		for _, instance := range instanceList.Items {
			for _, metadata := range instance.Metadata.Items {
				if strings.EqualFold(metadata.Key, "created-by") && strings.EqualFold(*metadata.Value, filter) {
					workers = append(workers, instance)
				}
			}
		}

		allowedDeletions := 0

		for _, worker := range workers {
			timestamp, err := time.Parse(time.RFC3339, worker.CreationTimestamp)
			if err != nil {
				logrus.WithError(err).Error("Failed to parse timestamp")
			}
			if time.Now().After(timestamp.Add(minimumAge)) {
				allowedDeletions++
			}
		}
		n = autoscaler.Max(currentSize-allowedDeletions, 0)
	}


	if n != currentSize {
		logrus.Infof("Setting target size to %d from %d", n, currentSize)
		_, err = p.service.InstanceGroupManagers.Resize(p.projectId, p.zone, p.instanceGroup, int64(n)).Do()
		if err != nil {
			logrus.WithError(err).Error("Failed to resize the instance group")
			return err
		}
	}
	return nil
}
