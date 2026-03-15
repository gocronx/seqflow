import React from 'react';
import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Layout from '@theme/Layout';

function Hero() {
  const { siteConfig } = useDocusaurusContext();
  return (
    <header style={{
      padding: '5rem 0 4rem',
      textAlign: 'center',
      background: 'linear-gradient(135deg, #0f172a 0%, #1e3a8a 50%, #2563eb 100%)',
      color: '#fff',
    }}>
      <div className="container">
        <p style={{
          fontSize: '0.875rem',
          textTransform: 'uppercase',
          letterSpacing: '3px',
          opacity: 0.7,
          marginBottom: '1rem',
        }}>
          序列驱动的 Disruptor
        </p>
        <h1 style={{
          fontSize: '4rem',
          fontWeight: 800,
          marginBottom: '1rem',
          letterSpacing: '-2px',
        }}>
          {siteConfig.title}
        </h1>
        <p style={{
          fontSize: '1.3rem',
          opacity: 0.9,
          marginBottom: '2.5rem',
          maxWidth: '600px',
          margin: '0 auto 2.5rem',
          lineHeight: 1.6,
        }}>
          高性能无锁 Disruptor，Go 实现
        </p>
        <div style={{ display: 'flex', gap: '1rem', justifyContent: 'center', flexWrap: 'wrap' }}>
          <Link className="button button--lg" to="/zh-Hans/docs/getting-started"
            style={{ background: '#fff', color: '#1e3a8a', fontWeight: 700, border: 'none' }}>
            快速开始
          </Link>
          <Link className="button button--lg"
            href="https://github.com/gocronx/seqflow"
            style={{ background: 'transparent', color: '#fff', border: '2px solid rgba(255,255,255,0.4)' }}>
            GitHub
          </Link>
        </div>
      </div>
    </header>
  );
}

function Features() {
  const features = [
    {
      title: '比 Channel 快 10 倍',
      desc: '单写者 Reserve 仅 2.1 ns/op。批量预留快 160 倍。零内存分配，零 GC。',
      color: '#2563eb',
    },
    {
      title: 'DAG 消费者拓扑',
      desc: '流水线、菱形、扇出。通过 DependsOn() 声明依赖。支持任意有向无环图。',
      color: '#3b82f6',
    },
    {
      title: '生产就绪',
      desc: '4 种等待策略，可选指标，优雅关闭。单包设计，无外部依赖。',
      color: '#60a5fa',
    },
  ];
  return (
    <section style={{ padding: '4rem 0', background: '#f8fafc' }}>
      <div className="container">
        <div className="row">
          {features.map((f, i) => (
            <div key={i} className="col col--4" style={{ marginBottom: '2rem' }}>
              <div style={{
                padding: '2rem',
                borderRadius: '12px',
                background: '#fff',
                border: '1px solid #e2e8f0',
                height: '100%',
                boxShadow: '0 1px 3px rgba(0,0,0,0.05)',
              }}>
                <div style={{
                  width: 8, height: 8, borderRadius: '50%',
                  background: f.color, marginBottom: '1rem',
                }} />
                <h3 style={{ fontSize: '1.15rem', marginBottom: '0.75rem' }}>{f.title}</h3>
                <p style={{ color: '#64748b', lineHeight: 1.6, marginBottom: 0 }}>{f.desc}</p>
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}

export default function Home() {
  return (
    <Layout title="首页" description="高性能无锁 Disruptor，Go 实现">
      <Hero />
      <Features />
    </Layout>
  );
}
