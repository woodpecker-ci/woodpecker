import humanizeDuration from "humanize-duration";
import React from "react";

export default class Duration extends React.Component {
	render() {
		const { start, finished } = this.props;

		return <time>{humanizeDuration((finished - start) * 1000)}</time>;
	}
}
