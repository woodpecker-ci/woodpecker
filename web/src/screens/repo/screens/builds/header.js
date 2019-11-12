import React, { Component } from "react";
import { Link } from "react-router-dom";
import Breadcrumb from "shared/components/breadcrumb";

export default class Header extends Component {
	render() {
		const { owner, repo } = this.props.match.params;
		return (
			<div>
				<Breadcrumb
					elements={[
						<Link to={`/${owner}/${repo}`} key={`${owner}-${repo}`}>
							{owner} / {repo}
						</Link>,
					]}
				/>
			</div>
		);
	}
}
