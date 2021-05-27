import React, { Component } from "react";
import { Link } from "react-router-dom";
import { List, Item } from "./components";

import { fetchBuildList, compareBuild } from "shared/utils/build";
import { fetchRepository, repositorySlug } from "shared/utils/repository";

import { branch } from "baobab-react/higher-order";
import { inject } from "config/client/inject";

import styles from "./index.less";

const binding = (props, context) => {
  const { owner, repo } = props.match.params;
  const slug = repositorySlug(owner, repo);
  return {
    repo: ["repos", "data", slug],
    builds: ["builds", "data", slug],
    loaded: ["builds", "loaded"],
    error: ["builds", "error"],
  };
};

@inject
@branch(binding)
export default class Main extends Component {
  constructor(props, context) {
    super(props, context);

    this.fetchNextBuildPage = this.fetchNextBuildPage.bind(this);
    this.selectBranch = this.selectBranch.bind(this)
  }

  componentWillMount() {
    this.synchronize(this.props);
  }

  shouldComponentUpdate(nextProps, nextState) {
    return (
      this.props.repo !== nextProps.repo ||
      (nextProps.builds !== undefined &&
        this.props.builds !== nextProps.builds) ||
      this.props.error !== nextProps.error ||
      this.props.loaded !== nextProps.loaded ||
      this.state.branch !== nextState.branch
    );
  }

  componentWillUpdate(nextProps) {
    if (this.props.match.url !== nextProps.match.url) {
      this.synchronize(nextProps);
    }
  }

  componentDidUpdate(prevProps) {
    if (this.props.location !== prevProps.location) {
      window.scrollTo(0, 0);
    }
  }

  synchronize(props) {
    const { drone, dispatch, match, repo } = props;

    if (!repo) {
      dispatch(fetchRepository, drone, match.params.owner, match.params.repo);
    }

    dispatch(fetchBuildList, drone, match.params.owner, match.params.repo);
  }

  fetchNextBuildPage(buildList) {
    const { drone, dispatch, match } = this.props;
    const page = Math.floor(buildList.length / 50) + 1;

    dispatch(
      fetchBuildList,
      drone,
      match.params.owner,
      match.params.repo,
      page,
    );
  }

  selectBranch(branch) {
    this.setState({
      branch: branch,
    });
  }

  render() {
    const { repo, builds, loaded, error } = this.props;
    const { branch } = this.state;
    const list = Object.values(builds || {});

    function renderBuild(build) {
      return (
        <Link to={`/${repo.full_name}/${build.number}`} key={build.number}>
          <Item build={build} />
        </Link>
      );
    }

    const filterBranch = (build) => {
      return !branch || build.branch === branch;
    }

    if (error) {
      return <div>Not Found</div>;
    }

    if (!loaded && list.length === 0) {
      return <div>Loading</div>;
    }

    if (!repo) {
      return <div>Loading</div>;
    }

    if (list.length === 0) {
      return <div>Build list is empty</div>;
    }

    return (
      <div className={styles.root}>
        <div className={styles.right}>
          {!branch ?
            <button onClick={() => this.selectBranch(repo.default_branch)}>Show {repo.default_branch} branch only</button> :
            <button onClick={() => this.selectBranch(undefined)}>Show all branches</button>
          }
        </div>
        <List>{list.sort(compareBuild).filter(filterBranch).map(renderBuild)}</List>
        {list.length < repo.last_build && (
          <button
            onClick={() => this.fetchNextBuildPage(list)}
            className={styles.more}
          >
            Show more builds
          </button>
        )}
      </div>
    );
  }
}
