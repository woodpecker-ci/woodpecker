import React, { Component } from "react";

export class Elapsed extends Component {
  constructor(props, context) {
    super(props);

    this.state = {
      elapsed: 0,
    };

    this.tick = this.tick.bind(this);
  }

  componentDidMount() {
    this.timer = setInterval(this.tick, 1000);
  }

  componentWillUnmount() {
    clearInterval(this.timer);
  }

  tick() {
    const { start } = this.props;
    const stop = ~~(Date.now() / 1000);
    this.setState({
      elapsed: stop - start,
    });
  }

  render() {
    const { elapsed } = this.state;
    const date = new Date(null);
    date.setSeconds(elapsed);
    return (
      <time>
        {!elapsed ? (
          undefined
        ) : elapsed > 3600 ? (
          date.toISOString().substr(11, 8)
        ) : (
          date.toISOString().substr(14, 5)
        )}
      </time>
    );
  }
}

/*
 * Returns the duration in hh:mm:ss format.
 *
 * @param {number} from - The start time in secnds
 * @param {number} to - The end time in seconds
 * @return {string}
 */
export const formatTime = (end, start) => {
  const diff = end - start;
  const date = new Date(null);
  date.setSeconds(diff);

  return diff > 3600
    ? date.toISOString().substr(11, 8)
    : date.toISOString().substr(14, 5);
};
