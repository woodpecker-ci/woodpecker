import React, { Component } from "react";
import classnames from "classnames";
import {
	STATUS_BLOCKED,
	STATUS_DECLINED,
	STATUS_ERROR,
	STATUS_FAILURE,
	STATUS_KILLED,
	STATUS_PENDING,
	STATUS_RUNNING,
	STATUS_SKIPPED,
	STATUS_STARTED,
	STATUS_SUCCESS,
} from "shared/constants/status";
import style from "./status.less";

import {
	CheckIcon,
	CloseIcon,
	ClockIcon,
	RefreshIcon,
	RemoveIcon,
} from "./icons/index";

const defaultIconSize = 15;

const statusLabel = status => {
	switch (status) {
		case STATUS_BLOCKED:
			return "Pending Approval";
		case STATUS_DECLINED:
			return "Declined";
		case STATUS_ERROR:
			return "Error";
		case STATUS_FAILURE:
			return "Failure";
		case STATUS_KILLED:
			return "Cancelled";
		case STATUS_PENDING:
			return "Pending";
		case STATUS_RUNNING:
			return "Running";
		case STATUS_SKIPPED:
			return "Skipped";
		case STATUS_STARTED:
			return "Running";
		case STATUS_SUCCESS:
			return "Successful";
		default:
			return "";
	}
};

const renderIcon = (status, size) => {
	switch (status) {
		case STATUS_SKIPPED:
			return <RemoveIcon size={size} />;
		case STATUS_PENDING:
			return <ClockIcon size={size} />;
		case STATUS_RUNNING:
		case STATUS_STARTED:
			return <RefreshIcon size={size} />;
		case STATUS_SUCCESS:
			return <CheckIcon size={size} />;
		default:
			return <CloseIcon size={size} />;
	}
};

export default class Status extends Component {
	shouldComponentUpdate(nextProps, nextState) {
		return this.props.status !== nextProps.status;
	}

	render() {
		const { status } = this.props;
		const icon = renderIcon(status, defaultIconSize);
		const classes = classnames(style.root, style[status]);
		return <div className={classes}>{icon}</div>;
	}
}

export const StatusLabel = ({ status }) => {
	return (
		<div className={classnames(style.label, style[status])}>
			<div>{statusLabel(status)}</div>
		</div>
	);
};

export const StatusText = ({ status, text }) => {
	return (
		<div
			className={classnames(style.label, style[status])}
			style="text-transform: capitalize;padding: 5px 10px;"
		>
			<div>{text}</div>
		</div>
	);
};
