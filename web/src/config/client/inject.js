import React from "react";

export const drone = (client, Component) => {
  // @see https://github.com/yannickcr/eslint-plugin-react/issues/512
  // eslint-disable-next-line react/display-name
  const component = class extends React.Component {
    getChildContext() {
      return {
        drone: client,
      };
    }

    render() {
      return <Component {...this.state} {...this.props} />;
    }
  };

  component.childContextTypes = {
    drone: (props, propName) => {},
  };

  return component;
};

export const inject = Component => {
  // @see https://github.com/yannickcr/eslint-plugin-react/issues/512
  // eslint-disable-next-line react/display-name
  const component = class extends React.Component {
    render() {
      this.props.drone = this.context.drone;
      return <Component {...this.state} {...this.props} />;
    }
  };

  return component;
};
