import React from 'react';
import clsx from 'clsx';
import styles from './HomepageFeatures.module.css';

const FeatureList = [
  {
    title: 'OpenSource and free',
    Svg: require('../../static/img/feat-opensource.svg').default,
    description: (
      <>
        Woodpecker is and always will be totally free. As Woodpecker's{' '}
        <a href="https://github.com/woodpecker-ci/woodpecker" target="_blank">
          source code
        </a>{' '}
        is open-source you can contribute to help evolving the project.
      </>
    ),
  },
  {
    title: 'Based on docker containers',
    Svg: require('../../static/img/feat-docker.svg').default,
    description: (
      <>
        Woodpecker uses docker containers to execute pipeline steps. If you need more than a normal docker image, you
        can create plugins to extend the pipeline features.{' '}
        <a href="/docs/usage/plugins/plugins">How do plugins work?</a>
      </>
    ),
  },
  {
    title: 'Multi workflows',
    Svg: require('../../static/img/feat-multipipelines.svg').default,
    description: (
      <>
        Woodpecker allows you to easily create multiple workflows for your project. They can even depend on each other.
        Check out the <a href="/docs/usage/workflows">docs</a>
      </>
    ),
  },
];

function Feature({ Svg, title, description }) {
  return (
    <div className={clsx('col col--4')}>
      <div className="text--center">
        <Svg className={styles.featureSvg} alt={title} />
      </div>
      <div className="text--center padding-horiz--md">
        <h3>{title}</h3>
        <p>{description}</p>
      </div>
    </div>
  );
}

export default function HomepageFeatures() {
  return (
    <section className={styles.features}>
      <div className="container">
        <div className="row">
          {FeatureList.map((props, idx) => (
            <Feature key={idx} {...props} />
          ))}
        </div>
      </div>
    </section>
  );
}
