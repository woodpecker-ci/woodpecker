import React, { Component } from "react";

import { branch } from "baobab-react/higher-order";
import { inject } from "config/client/inject";

import {
  fetchRepostoryList,
  disableRepository,
  enableRepository
} from "shared/utils/repository";

import { List, Item } from "./components";
import Breadcrumb, { SEPARATOR } from "shared/components/breadcrumb";

import styles from "./index.less";

const binding = (props, context) => {
  return {
    repos: ["repos", "data"],
    loaded: ["repos", "loaded"],
    error: ["repos", "error"]
  };
};

@inject
@branch(binding)
export default class UserRepos extends Component {
  constructor(props, context) {
    super(props, context);

    this.handleFilter = this.handleFilter.bind(this);
    this.renderItem = this.renderItem.bind(this);
    this.handleToggle = this.handleToggle.bind(this);
  }

  handleFilter(e) {
    this.setState({
      search: e.target.value
    });
  }

  handleToggle(repo, e) {
    const { dispatch, drone } = this.props;
    if (e.target.checked) {
      dispatch(enableRepository, drone, repo.owner, repo.name);
    } else {
      dispatch(disableRepository, drone, repo.owner, repo.name);
    }
  }

  componentWillMount() {
    if (!this._dispatched) {
      this._dispatched = true;
      this.props.dispatch(fetchRepostoryList, this.props.drone);
    }
  }

  shouldComponentUpdate(nextProps, nextState) {
    return (
      this.props.repos !== nextProps.repos ||
      this.state.search !== nextState.search
    );
  }

  render() {
    const { repos, loaded, error } = this.props;
    const { search } = this.state;
    const list = Object.values(repos || {});

    if (error) {
      return ERROR;
    }

    if (!loaded) {
      return LOADING;
    }

    if (list.length === 0) {
      return EMPTY;
    }

    const filter = repo => {
      return !search || repo.full_name.indexOf(search) !== -1;
    };

    const filtered = list.filter(filter);

    return (
      <div>
        <div className={styles.search}>
          <input
            type="text"
            placeholder="Search …"
            onChange={this.handleFilter}
          />
        </div>
        <div className={styles.root}>
          {filtered.length === 0 ? NO_MATCHES : null}
          <List>{list.filter(filter).map(this.renderItem)}</List>
        </div>
      </div>
    );
  }

  renderItem(repo) {
    return (
      <Item
        key={repo.full_name}
        owner={repo.owner}
        name={repo.name}
        active={repo.active}
        link={`/${repo.full_name}`}
        onchange={this.handleToggle.bind(this, repo)}
      />
    );
  }
}

const LOADING = <div>Loading</div>;

const EMPTY = <div>Your repository list is empty</div>;

const NO_MATCHES = <div>No matches found</div>;

const ERROR = <div>Error</div>;

/* eslint-disable react/jsx-key */
export class UserRepoTitle extends Component {
  render() {
    return (
      <Breadcrumb
        elements={[<span>Account</span>, SEPARATOR, <span>Repositories</span>]}
      />
    );
  }
}
/* eslint-enable react/jsx-key */
