import clsx from "clsx";
import Link from "@docusaurus/Link";
import useDocusaurusContext from "@docusaurus/useDocusaurusContext";
import Layout from "@theme/Layout";
import HomepageFeatures from "@site/src/components/HomepageFeatures";
import styles from "./index.module.css";

function HomepageHeader() {
  const { siteConfig } = useDocusaurusContext();
  return (
    <header className={clsx("hero hero--primary", styles.heroBanner)}>
      <div className="container">
        <h1 className="hero__title">{siteConfig.title}</h1>
        <p className="hero__subtitle">
          Обход блокировок на уровне сетевых пакетов для Linux и роутеров
        </p>
        <p className={styles.description}>
          Веб-интерфейс, автоматический подбор конфигурации, поддержка роутеров
          и серверов
        </p>
        <div className={styles.buttons}>
          <Link
            className="button button--secondary button--lg"
            to="/docs/intro"
          >
            Начать
          </Link>
          <Link
            className="button button--outline button--secondary button--lg margin-left--md"
            to="https://github.com/DanielLavrushin/b4"
          >
            GitHub
          </Link>
        </div>
      </div>
    </header>
  );
}

export default function Home() {
  return (
    <Layout title="Документация" description="B4 — обход блокировок для Linux">
      <HomepageHeader />
      <main>
        <HomepageFeatures />
      </main>
    </Layout>
  );
}
