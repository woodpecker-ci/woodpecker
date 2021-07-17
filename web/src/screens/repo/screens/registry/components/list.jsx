import React from "react";
import styles from "./list.less";

export const List = ({ children }) => (
  <div className={styles.list}>{children}</div>
);

export const Item = props => (
  <div className={styles.item} key={props.name}>
    <div>{props.name}</div>
    <div>
      <button onClick={props.ondelete}>delete</button>
    </div>
  </div>
);
