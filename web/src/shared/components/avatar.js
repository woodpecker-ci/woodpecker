import React, { Component } from "react";
import styles from "./avatar.less";

export default class Avatar extends Component {
	render() {
		const image = this.props.image;
		const style = {
			backgroundImage: `url(${image})`,
		};
		return <div className={styles.avatar} style={style} />;
	}
}
