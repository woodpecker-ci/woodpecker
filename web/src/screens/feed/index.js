import React, { Component } from "react";
import { Link } from "react-router-dom";

import { compareFeedItem } from "shared/utils/feed";

import { branch } from "baobab-react/higher-order";
import { inject } from "config/client/inject";

import DroneIcon from "shared/components/logo";
import { List, Item } from "./components";

import style from "./index.less";

import Collapsible from "react-collapsible";

const binding = (props, context) => {
  return { feed: ["feed"] };
};

@inject
@branch(binding)
export default class Sidebar extends Component {
  constructor(props, context) {
    super(props, context);

    this.setState({
      starred: JSON.parse(localStorage.getItem("starred") || "[]"),
      starredOpen: (localStorage.getItem("starredOpen") || "true") === "true",
      reposOpen: (localStorage.getItem("reposOpen") || "true") === "true",
    });

    this.handleFilter = this.handleFilter.bind(this);
    this.toggleStarred = this.toggleItem.bind(this, "starredOpen");
    this.toggleAll = this.toggleItem.bind(this, "reposOpen");
  }

  shouldComponentUpdate(nextProps, nextState) {
    return (
      this.props.feed !== nextProps.feed ||
      this.state.filter !== nextState.filter ||
      this.state.starred.length !== nextState.starred.length
    );
  }

  handleFilter(e) {
    this.setState({
      filter: e.target.value,
    });
  }

  toggleItem = item => {
    this.setState(state => {
      return { [item]: !state[item] };
    });

    localStorage.setItem(item, this.state[item]);
  };

  renderFeed = (list, renderStarred) => {
    return (
      <div>
        <List>{list.map(item => this.renderItem(item, renderStarred))}</List>
      </div>
    );
  };

  renderItem = (item, renderStarred) => {
    const starred = this.state.starred;
    if (renderStarred && !starred.includes(item.full_name)) {
      return null;
    }
    return (
      <Link to={`/${item.full_name}`} key={item.full_name}>
        <Item
          item={item}
          onFave={this.onFave}
          faved={starred.includes(item.full_name)}
        />
      </Link>
    );
  };

  onFave = fullName => {
    if (!this.state.starred.includes(fullName)) {
      this.setState(state => {
        const list = state.starred.concat(fullName);
        return { starred: list };
      });
    } else {
      this.setState(state => {
        const list = state.starred.filter(v => v !== fullName);
        return { starred: list };
      });
    }

    localStorage.setItem("starred", JSON.stringify(this.state.starred));
  };

  render() {
    const { feed } = this.props;
    const { filter } = this.state;

    const list = feed.data ? Object.values(feed.data) : [];

    const filterFunc = item => {
      return !filter || item.full_name.indexOf(filter) !== -1;
    };

    const filtered = list.filter(filterFunc).sort(compareFeedItem);
    const starredOpen = this.state.starredOpen;
    const reposOpen = this.state.reposOpen;
    return (
      <div className={style.feed}>
        {LOGO}
        <Collapsible
          trigger="Starred"
          triggerTagName="div"
          transitionTime={200}
          open={starredOpen}
          onOpen={this.toggleStarred}
          onClose={this.toggleStarred}
          triggerOpenedClassName={style.Collapsible__trigger}
          triggerClassName={style.Collapsible__trigger}
        >
          {feed.loaded === false ? (
            LOADING
          ) : feed.error ? (
            ERROR
          ) : list.length === 0 ? (
            EMPTY
          ) : (
            this.renderFeed(list, true)
          )}
        </Collapsible>
        <Collapsible
          trigger="Repos"
          triggerTagName="div"
          transitionTime={200}
          open={reposOpen}
          onOpen={this.toggleAll}
          onClose={this.toggleAll}
          triggerOpenedClassName={style.Collapsible__trigger}
          triggerClassName={style.Collapsible__trigger}
        >
          <input
            type="text"
            placeholder="Search …"
            onChange={this.handleFilter}
          />
          {feed.loaded === false ? (
            LOADING
          ) : feed.error ? (
            ERROR
          ) : list.length === 0 ? (
            EMPTY
          ) : filtered.length > 0 ? (
            this.renderFeed(filtered.sort(compareFeedItem), false)
          ) : (
            NO_MATCHES
          )}
        </Collapsible>
      </div>
    );
  }
}

const LOGO = (
  <div className={style.brand}>
    <DroneIcon />
    <p>
      Woodpecker<span style="margin-left: 4px;">
        {window.WOODPECKER_VERSION}
      </span>
      <br />
      <span>
        <a
          href="{window.WOODPECKER_DOCS}"
          target="_blank"
          rel="noopener noreferrer"
        >
          Docs
        </a>
      </span>
    </p>
  </div>
);

const LOADING = <div className={style.message}>Loading</div>;

const EMPTY = <div className={style.message}>Your build feed is empty</div>;

const NO_MATCHES = <div className={style.message}>No results found</div>;

const ERROR = (
  <div className={style.message}>
    Oops. It looks like there was a problem loading your feed
  </div>
);
