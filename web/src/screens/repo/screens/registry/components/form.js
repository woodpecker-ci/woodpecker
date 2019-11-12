import React, { Component } from "react";
import styles from "./form.less";

export class Form extends Component {
	constructor(props, context) {
		super(props, context);

		this.state = {
			address: "",
			username: "",
			password: "",
		};

		this._handleAddressChange = this._handleAddressChange.bind(this);
		this._handleUsernameChange = this._handleUsernameChange.bind(this);
		this._handlePasswordChange = this._handlePasswordChange.bind(this);
		this._handleSubmit = this._handleSubmit.bind(this);

		this.clear = this.clear.bind(this);
	}

	_handleAddressChange(event) {
		this.setState({ address: event.target.value });
	}

	_handleUsernameChange(event) {
		this.setState({ username: event.target.value });
	}

	_handlePasswordChange(event) {
		this.setState({ password: event.target.value });
	}

	_handleSubmit() {
		const { onsubmit } = this.props;

		const detail = {
			address: this.state.address,
			username: this.state.username,
			password: this.state.password,
		};

		onsubmit({ detail });
		this.clear();
	}

	clear() {
		this.setState({ address: "" });
		this.setState({ username: "" });
		this.setState({ password: "" });
	}

	render() {
		return (
			<div className={styles.form}>
				<input
					type="text"
					value={this.state.address}
					onChange={this._handleAddressChange}
					placeholder="Registry Address (e.g. docker.io)"
				/>
				<input
					type="text"
					value={this.state.username}
					onChange={this._handleUsernameChange}
					placeholder="Registry Username"
				/>
				<textarea
					rows="1"
					value={this.state.password}
					onChange={this._handlePasswordChange}
					placeholder="Registry Password"
				/>
				<div className={styles.actions}>
					<button onClick={this._handleSubmit}>Save</button>
				</div>
			</div>
		);
	}
}
