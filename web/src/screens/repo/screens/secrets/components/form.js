import React, { Component } from "react";

import {
  EVENT_PUSH,
  EVENT_TAG,
  EVENT_PULL_REQUEST,
  EVENT_DEPLOY,
} from "shared/constants/events";

import styles from "./form.less";

export class Form extends Component {
  constructor(props, context) {
    super(props, context);

    this.state = {
      name: "",
      value: "",
      event: [EVENT_PUSH, EVENT_TAG, EVENT_DEPLOY],
    };

    this._handleNameChange = this._handleNameChange.bind(this);
    this._handleValueChange = this._handleValueChange.bind(this);
    this._handleEventChange = this._handleEventChange.bind(this);
    this._handleSubmit = this._handleSubmit.bind(this);

    this.clear = this.clear.bind(this);
  }

  _handleNameChange(event) {
    this.setState({ name: event.target.value });
  }

  _handleValueChange(event) {
    this.setState({ value: event.target.value });
  }

  _handleEventChange(event) {
    const selected = this.state.event;
    let index;

    if (event.target.checked) {
      selected.push(event.target.value);
    } else {
      index = selected.indexOf(event.target.value);
      selected.splice(index, 1);
    }

    this.setState({ event: selected });
  }

  _handleSubmit() {
    const { onsubmit } = this.props;

    const detail = {
      name: this.state.name,
      value: this.state.value,
      event: this.state.event,
    };

    onsubmit({ detail });
    this.clear();
  }

  clear() {
    this.setState({ name: "" });
    this.setState({ value: "" });
    this.setState({ event: [EVENT_PUSH, EVENT_TAG, EVENT_DEPLOY] });
  }

  render() {
    let checked = this.state.event.reduce((map, event) => {
      map[event] = true;
      return map;
    }, {});

    return (
      <div className={styles.form}>
        <input
          type="text"
          name="name"
          value={this.state.name}
          placeholder="Secret Name"
          onChange={this._handleNameChange}
        />
        <textarea
          rows="1"
          name="value"
          value={this.state.value}
          placeholder="Secret Value"
          onChange={this._handleValueChange}
        />
        <section>
          <h2>Events</h2>
          <div>
            <label>
              <input
                type="checkbox"
                checked={checked[EVENT_PUSH]}
                value={EVENT_PUSH}
                onChange={this._handleEventChange}
              />
              <span>push</span>
            </label>
            <label>
              <input
                type="checkbox"
                checked={checked[EVENT_TAG]}
                value={EVENT_TAG}
                onChange={this._handleEventChange}
              />
              <span>tag</span>
            </label>
            <label>
              <input
                type="checkbox"
                checked={checked[EVENT_PULL_REQUEST]}
                value={EVENT_PULL_REQUEST}
                onChange={this._handleEventChange}
              />
              <span>pull request</span>
            </label>
            <label>
              <input
                type="checkbox"
                checked={checked[EVENT_DEPLOY]}
                value={EVENT_DEPLOY}
                onChange={this._handleEventChange}
              />
              <span>deploy</span>
            </label>
          </div>
        </section>
        <div className={styles.actions}>
          <button onClick={this._handleSubmit}>Save</button>
        </div>
      </div>
    );
  }
}
