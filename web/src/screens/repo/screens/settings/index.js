import React, { Component } from "react";

import { branch } from "baobab-react/higher-order";
import { inject } from "config/client/inject";

import {
  fetchRepository,
  updateRepository,
  repositorySlug,
} from "shared/utils/repository";

import {
  VISIBILITY_PUBLIC,
  VISIBILITY_PRIVATE,
  VISIBILITY_INTERNAL,
} from "shared/constants/visibility";

import styles from "./index.less";

const binding = (props, context) => {
  const { owner, repo } = props.match.params;
  const slug = repositorySlug(owner, repo);
  return {
    user: ["user", "data"],
    repo: ["repos", "data", slug],
  };
};

@inject
@branch(binding)
export default class Settings extends Component {
  constructor(props, context) {
    super(props, context);

    this.handlePushChange = this.handlePushChange.bind(this);
    this.handlePullChange = this.handlePullChange.bind(this);
    this.handleTagChange = this.handleTagChange.bind(this);
    this.handleDeployChange = this.handleDeployChange.bind(this);
    this.handleTrustedChange = this.handleTrustedChange.bind(this);
    this.handleProtectedChange = this.handleProtectedChange.bind(this);
    this.handleVisibilityChange = this.handleVisibilityChange.bind(this);
    this.handleTimeoutChange = this.handleTimeoutChange.bind(this);
    this.handlePathChange = this.handlePathChange.bind(this);
    this.handleChange = this.handleChange.bind(this);
  }

  shouldComponentUpdate(nextProps, nextState) {
    return this.props.repo !== nextProps.repo;
  }

  componentWillMount() {
    const { drone, dispatch, match, repo } = this.props;

    if (!repo) {
      dispatch(fetchRepository, drone, match.params.owner, match.params.repo);
    }
  }

  render() {
    const { repo } = this.props;

    if (!repo) {
      return undefined;
    }

    return (
      <div className={styles.root}>
        <section>
          <h2>Pipeline Path</h2>
          <div>
            <input
              type="text"
              value={repo.config_file}
              onBlur={this.handlePathChange}
            />
          </div>
        </section>
        <section>
          <h2>Repository Hooks</h2>
          <div>
            <label>
              <input
                type="checkbox"
                checked={repo.allow_push}
                onChange={this.handlePushChange}
              />
              <span>push</span>
            </label>
            <label>
              <input
                type="checkbox"
                checked={repo.allow_pr}
                onChange={this.handlePullChange}
              />
              <span>pull request</span>
            </label>
            <label>
              <input
                type="checkbox"
                checked={repo.allow_tags}
                onChange={this.handleTagChange}
              />
              <span>tag</span>
            </label>
            <label>
              <input
                type="checkbox"
                checked={repo.allow_deploys}
                onChange={this.handleDeployChange}
              />
              <span>deployment</span>
            </label>
          </div>
        </section>

        <section>
          <h2>Project Settings</h2>
          <div>
            <label>
              <input
                type="checkbox"
                checked={repo.gated}
                onChange={this.handleProtectedChange}
              />
              <span>Protected</span>
            </label>
            <label>
              <input
                type="checkbox"
                checked={repo.trusted}
                onChange={this.handleTrustedChange}
              />
              <span>Trusted</span>
            </label>
          </div>
        </section>

        <section>
          <h2>Project Visibility</h2>
          <div>
            <label>
              <input
                type="radio"
                name="visibility"
                value="public"
                checked={repo.visibility === VISIBILITY_PUBLIC}
                onChange={this.handleVisibilityChange}
              />
              <span>Public</span>
            </label>
            <label>
              <input
                type="radio"
                name="visibility"
                value="private"
                checked={repo.visibility === VISIBILITY_PRIVATE}
                onChange={this.handleVisibilityChange}
              />
              <span>Private</span>
            </label>
            <label>
              <input
                type="radio"
                name="visibility"
                value="internal"
                checked={repo.visibility === VISIBILITY_INTERNAL}
                onChange={this.handleVisibilityChange}
              />
              <span>Internal</span>
            </label>
          </div>
        </section>

        <section>
          <h2>Timeout</h2>
          <div>
            <input
              type="number"
              value={repo.timeout}
              onBlur={this.handleTimeoutChange}
            />
            <span className={styles.minutes}>minutes</span>
          </div>
        </section>
      </div>
    );
  }

  handlePushChange(e) {
    this.handleChange("allow_push", e.target.checked);
  }

  handlePullChange(e) {
    this.handleChange("allow_pr", e.target.checked);
  }

  handleTagChange(e) {
    this.handleChange("allow_tag", e.target.checked);
  }

  handleDeployChange(e) {
    this.handleChange("allow_deploy", e.target.checked);
  }

  handleTrustedChange(e) {
    this.handleChange("trusted", e.target.checked);
  }

  handleProtectedChange(e) {
    this.handleChange("gated", e.target.checked);
  }

  handleVisibilityChange(e) {
    this.handleChange("visibility", e.target.value);
  }

  handleTimeoutChange(e) {
    this.handleChange("timeout", parseInt(e.target.value));
  }

  handlePathChange(e) {
    this.handleChange("config_file", e.target.value);
  }

  handleChange(prop, value) {
    const { dispatch, drone, repo } = this.props;
    let data = {};
    data[prop] = value;
    dispatch(updateRepository, drone, repo.owner, repo.name, data);
  }
}
