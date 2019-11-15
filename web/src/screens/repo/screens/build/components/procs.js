import React, { Component } from "react";
import { Link } from "react-router-dom";
import classnames from "classnames";

import { Elapsed, formatTime } from "./elapsed";
import { default as Status, StatusText } from "shared/components/status";

import styles from "./procs.less";

const renderEnviron = data => {
  return (
    <div>
      {data[0]}={data[1]}
    </div>
  );
};

const ProcListHolder = ({ vars, renderName, children }) => (
  <div className={styles.list}>
    {renderName && vars.name !== "drone" ? (
      <div>
        <StatusText status={vars.state} text={vars.name} />
      </div>
    ) : null}
    {vars.environ ? (
      <div>
        <StatusText
          status={vars.state}
          text={Object.entries(vars.environ).map(renderEnviron)}
        />
      </div>
    ) : null}
    {children}
  </div>
);

export class ProcList extends Component {
  render() {
    const { repo, build, rootProc, selectedProc, renderName } = this.props;
    return (
      <ProcListHolder vars={rootProc} renderName={renderName}>
        {this.props.rootProc.children.map(function(child) {
          return (
            <Link
              to={`/${repo.full_name}/${build.number}/${child.pid}`}
              key={`${repo.full_name}-${build.number}-${child.pid}`}
            >
              <ProcListItem
                key={child.pid}
                name={child.name}
                start={child.start_time}
                finish={child.end_time}
                state={child.state}
                selected={child.pid === selectedProc.pid}
              />
            </Link>
          );
        })}
      </ProcListHolder>
    );
  }
}

export const ProcListItem = ({ name, start, finish, state, selected }) => (
  <div className={classnames(styles.item, selected ? styles.selected : null)}>
    <h3>{name}</h3>
    {finish ? (
      <time>{formatTime(finish, start)}</time>
    ) : (
      <Elapsed start={start} />
    )}
    <div>
      <Status status={state} />
    </div>
  </div>
);
