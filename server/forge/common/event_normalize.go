package common

func NormalizeEventReason(in string) string {
	switch in {
	case "labels_cleared":
		return "label_cleared"
	case "labels_updated":
		return "label_updated"
	case "labels_added":
		return "label_added"
	}
	return in
}
