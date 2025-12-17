import React, { useState, useEffect, useRef } from 'react';
import Sidebar from '../components/Sidebar';
import Header from '../components/Header';
import { useI18n } from '../i18n';
import type { Experiment } from '../types';

const PluginPreviewPage: React.FC = () => {
  const { t } = useI18n();
  const iframeRef = useRef<HTMLIFrameElement>(null);

  // 配置项
  const [widgetSrc, setWidgetSrc] = useState('/experiment-widget.js');
  const [apiBase, setApiBase] = useState('http://localhost:4000/api/v1');
  const [experimentId, setExperimentId] = useState('');
  const [userKey, setUserKey] = useState('demo-user-' + Math.random().toString(36).substring(7));
  const [randomAssignment, setRandomAssignment] = useState(true);

  // 状态
  const [activeExperiments, setActiveExperiments] = useState<Experiment[]>([]);
  const [loadingExperiments, setLoadingExperiments] = useState(false);
  const [experimentsLoadError, setExperimentsLoadError] = useState<string>('');
  const [previewKey, setPreviewKey] = useState(0);

  useEffect(() => {
    loadActiveExperiments();
  }, [apiBase]);

  const loadActiveExperiments = async () => {
    setLoadingExperiments(true);
    setExperimentsLoadError('');

    const url = `${apiBase}/experiments?page=1&page_size=100&status=active`;
    console.log('🔄 [Plugin Preview] 开始加载实验列表...');
    console.log('📍 [Plugin Preview] API Base:', apiBase);
    console.log('🌐 [Plugin Preview] 请求 URL:', url);

    try {
      const startTime = Date.now();
      const response = await fetch(url, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
      });
      const elapsed = Date.now() - startTime;

      console.log(`📥 [Plugin Preview] 响应状态: ${response.status} ${response.statusText} (耗时: ${elapsed}ms)`);
      console.log('📋 [Plugin Preview] 响应头:', {
        'content-type': response.headers.get('content-type'),
        'access-control-allow-origin': response.headers.get('access-control-allow-origin'),
      });

      if (!response.ok) {
        const errorText = await response.text();
        console.error('❌ [Plugin Preview] HTTP 错误响应:', errorText);
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      const res = await response.json();
      console.log('📦 [Plugin Preview] 响应数据:', res);

      if (res.code === 0 && res.data) {
        const exps = res.data.experiments || [];
        console.log(`✅ [Plugin Preview] 成功加载 ${exps.length} 个激活的实验`);

        if (exps.length > 0) {
          console.log('📝 [Plugin Preview] 实验列表:', exps.map((e: any) => ({ id: e.experiment_id, name: e.name })));
        }

        setActiveExperiments(exps);
        if (exps.length > 0 && !experimentId) {
          setExperimentId(exps[0].experiment_id);
          console.log('🎯 [Plugin Preview] 自动选择第一个实验:', exps[0].experiment_id);
        }
      } else {
        console.error('❌ [Plugin Preview] API 返回错误:', res);
        throw new Error(res.message || '加载实验失败');
      }
    } catch (error: any) {
      console.error('❌ [Plugin Preview] 加载失败:', error);
      console.error('🔍 [Plugin Preview] 错误详情:', {
        message: error.message,
        stack: error.stack,
        type: error.name,
      });

      setActiveExperiments([]);

      // 提供更友好的错误信息
      let errorMsg = error.message || '网络错误';
      if (error.message?.includes('Failed to fetch')) {
        errorMsg = '网络连接失败，请检查后端服务是否运行';
      } else if (error.message?.includes('NetworkError')) {
        errorMsg = 'CORS 错误或网络问题';
      }
      setExperimentsLoadError(errorMsg);
    } finally {
      setLoadingExperiments(false);
      console.log('🏁 [Plugin Preview] 加载实验列表完成');
    }
  };

  const generatePreviewHTML = () => {
    return `<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Experiment Widget Preview</title>
  <style>
    * { margin: 0; padding: 0; box-sizing: border-box; }
    body {
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
      padding: 20px;
      background: #f5f5f5;
      display: flex;
      flex-direction: column;
      align-items: center;
      gap: 20px;
    }
    .info-panel {
      background: white;
      border-radius: 8px;
      padding: 16px;
      box-shadow: 0 1px 3px rgba(0,0,0,0.1);
      width: 100%;
      max-width: 800px;
    }
    .info-panel h3 {
      font-size: 14px;
      color: #333;
      margin-bottom: 12px;
      font-weight: 600;
    }
    .info-row {
      display: flex;
      align-items: flex-start;
      margin-bottom: 8px;
      font-size: 13px;
      line-height: 1.6;
    }
    .info-label {
      color: #666;
      min-width: 120px;
      font-weight: 500;
    }
    .info-value {
      color: #333;
      word-break: break-all;
      flex: 1;
    }
    .info-value code {
      background: #f5f5f5;
      padding: 2px 6px;
      border-radius: 3px;
      font-size: 12px;
      font-family: 'Monaco', 'Menlo', monospace;
    }
    .preview-area {
      background: white;
      border-radius: 8px;
      padding: 20px;
      min-height: 200px;
      box-shadow: 0 1px 3px rgba(0,0,0,0.1);
      width: 100%;
      max-width: 800px;
      display: flex;
      flex-direction: column;
      align-items: center;
    }
    .preview-title {
      font-size: 14px;
      color: #666;
      margin-bottom: 16px;
      padding-bottom: 12px;
      border-bottom: 1px solid #e8e8e8;
      width: 100%;
      text-align: center;
    }
    /* Widget 容器居中显示 */
    #exp-widget-root {
      display: flex;
      justify-content: center;
      align-items: flex-start;
      width: 100%;
    }
  </style>
</head>
<body>
  <div class="info-panel">
    <h3>${t('experimentConfig')}</h3>
    <div class="info-row">
      <span class="info-label">${t('widgetScriptUrl')}:</span>
      <span class="info-value"><code>${widgetSrc}</code></span>
    </div>
    <div class="info-row">
      <span class="info-label">${t('apiBaseUrl')}:</span>
      <span class="info-value"><code>${apiBase}</code></span>
    </div>
    <div class="info-row">
      <span class="info-label">${t('experimentId')}:</span>
      <span class="info-value"><code>${experimentId}</code></span>
    </div>
    <div class="info-row">
      <span class="info-label">${t('userKey')}:</span>
      <span class="info-value"><code>${userKey}</code></span>
    </div>
    <div class="info-row">
      <span class="info-label">${t('randomAssignment')}:</span>
      <span class="info-value"><code>${randomAssignment}</code></span>
    </div>
  </div>

  <div class="preview-area">
    <div class="preview-title">${t('widgetPreviewArea')}</div>
    <!-- SDK 会自动在 body 中创建 #exp-widget-root 元素 -->
  </div>

  <!--
    Experiment Widget SDK
    按照文档要求，只需要一个 script 标签即可
  -->
  <script
    src="${widgetSrc}"
    data-api-base="${apiBase}"
    data-experiment-id="${experimentId}"
    data-user-key="${userKey}"
    data-random-assignment="${randomAssignment}"
    async
  ></script>

  <!-- 监听插件事件（可选） -->
  <script>
    console.log('🎬 [Widget Preview] 预览环境初始化开始...');
    console.log('📍 [Widget Preview] Widget Script URL:', '${widgetSrc}');
    console.log('🌐 [Widget Preview] API Base:', '${apiBase}');
    console.log('🎯 [Widget Preview] Experiment ID:', '${experimentId}');
    console.log('👤 [Widget Preview] User Key:', '${userKey}');
    console.log('🎲 [Widget Preview] Random Assignment:', ${randomAssignment});

    // 监听 script 加载成功
    const scripts = document.querySelectorAll('script[src]');
    scripts.forEach(function(script) {
      script.addEventListener('load', function() {
        console.log('✅ [Widget Preview] 脚本加载成功:', script.src);
      });
      script.addEventListener('error', function(e) {
        console.error('❌ [Widget Preview] 脚本加载失败:', script.src);
        console.error('🔍 [Widget Preview] 错误详情:', e);
      });
    });

    // 监听插件事件
    window.addEventListener('message', function(event) {
      if (event.data && event.data.type === 'experiment-widget') {
        console.log('📨 [Widget Preview] Widget 事件:', event.data);
      }
    });

    // 捕获所有错误
    window.addEventListener('error', function(e) {
      console.error('❌ [Widget Preview] 运行时错误:');
      console.error('  消息:', e.message);
      console.error('  文件:', e.filename);
      console.error('  行号:', e.lineno, '列号:', e.colno);
      console.error('  错误对象:', e.error);

      if (e.filename && e.filename.includes('${widgetSrc}')) {
        console.error('⚠️ [Widget Preview] Widget 脚本执行出错！');
      }
    }, true);

    // 捕获 fetch 错误
    const originalFetch = window.fetch;
    window.fetch = function(...args) {
      const url = args[0];
      console.log('🌐 [Widget Preview] Widget 发起请求:', url);

      return originalFetch.apply(this, args)
        .then(function(response) {
          console.log(\`📥 [Widget Preview] 请求响应: \${response.status} - \${url}\`);
          return response;
        })
        .catch(function(error) {
          console.error('❌ [Widget Preview] 请求失败:', url, error);
          throw error;
        });
    };

    console.log('✅ [Widget Preview] 预览环境初始化完成，等待 Widget 加载...');
  </script>
</body>
</html>`;
  };

  const loadPreview = () => {
    console.log('🚀 [Plugin Preview] 用户点击"加载预览"按钮');

    if (!experimentId) {
      console.warn('⚠️ [Plugin Preview] 未选择实验，无法加载预览');
      alert(t('pleaseSelectExperimentFirst'));
      return;
    }

    console.log('📋 [Plugin Preview] 预览配置:');
    console.log('  Widget Script:', widgetSrc);
    console.log('  API Base:', apiBase);
    console.log('  Experiment ID:', experimentId);
    console.log('  User Key:', userKey);
    console.log('  Random Assignment:', randomAssignment);

    // 强制重新加载 iframe
    const newKey = Date.now();
    setPreviewKey(newKey);
    console.log(`🔄 [Plugin Preview] 重新加载 iframe (key: ${newKey})`);
  };

  const copyIntegrationCode = () => {
    const code = `<!-- Experiment Widget SDK -->
<script
  src="${widgetSrc}"
  data-api-base="${apiBase}"
  data-experiment-id="${experimentId || 'YOUR_EXPERIMENT_ID'}"
  data-user-key="${userKey || 'OPTIONAL_USER_KEY'}"
  data-random-assignment="${randomAssignment}"
  async>
</script>`;
    navigator.clipboard.writeText(code);
    alert(t('integrationCodeCopied'));
  };

  return (
    <div className="app">
      <Sidebar />

      <div className="main-content">
        <Header title={t('resourcesTitle')} />

        <div className="content">
          <div className="compact-layout">
            {/* 配置区域 */}
            <div className="compact-card">
              <div className="compact-card-header">
                <h3 className="compact-card-title">{t('widgetConfig')}</h3>
                <div className="compact-card-hint">{t('widgetConfigHint')}</div>
              </div>

              <div className="compact-card-body">
                <div className="compact-form-grid">
                  {/* Widget Script URL */}
                  <div className="compact-form-group">
                    <label className="compact-label">
                      <span className="label-text">{t('widgetScriptUrl')}</span>
                      <span className="label-required">*</span>
                    </label>
                    <input
                      className="compact-input"
                      value={widgetSrc}
                      onChange={(e) => setWidgetSrc(e.target.value)}
                      placeholder="https://experiment-widget-sdk.vercel.app/experiment-widget.xxx.js"
                    />
                    <div style={{ marginTop: 6, fontSize: 11, color: '#8c8c8c' }}>
                      💡 {t('widgetScriptHint')} <code style={{ background: '#f5f5f5', padding: '2px 6px' }}>/experiment-widget.js</code>
                    </div>
                  </div>

                  {/* API Base URL */}
                  <div className="compact-form-group">
                    <label className="compact-label">
                      <span className="label-text">{t('apiBaseUrl')}</span>
                      <span className="label-required">*</span>
                    </label>
                    <input
                      className="compact-input"
                      value={apiBase}
                      onChange={(e) => setApiBase(e.target.value)}
                      placeholder="http://localhost:4000/api/v1"
                    />
                    <div style={{ marginTop: 6, fontSize: 11, color: '#8c8c8c' }}>
                      {t('quickSwitch')}
                      <button
                        className="compact-btn compact-btn-text compact-btn-xs"
                        onClick={() => setApiBase('http://localhost:4000/api/v1')}
                        style={{ marginLeft: 4, padding: '2px 6px' }}
                      >
                        {t('localDev')}
                      </button>
                      <button
                        className="compact-btn compact-btn-text compact-btn-xs"
                        onClick={() => setApiBase('https://ads-creative-gen-platform-production.up.railway.app/api/v1')}
                        style={{ marginLeft: 4, padding: '2px 6px' }}
                      >
                        {t('production')}
                      </button>
                    </div>
                  </div>

                  {/* 实验选择 */}
                  <div className="compact-form-group">
                    <label className="compact-label">
                      <span className="label-text">{t('selectActiveExperiment')}</span>
                      <span className="label-required">*</span>
                      <button
                        className="compact-btn compact-btn-text compact-btn-xs"
                        onClick={loadActiveExperiments}
                        disabled={loadingExperiments}
                        style={{ marginLeft: 8, padding: '2px 8px' }}
                      >
                        <i className="fas fa-sync-alt" style={{ fontSize: 10 }}></i>
                        <span style={{ marginLeft: 4 }}>{t('refresh')}</span>
                      </button>
                    </label>

                    <select
                      className="compact-input"
                      value={experimentId}
                      onChange={(e) => setExperimentId(e.target.value)}
                      disabled={loadingExperiments}
                    >
                      <option value="">
                        {loadingExperiments ? t('loading') : activeExperiments.length === 0 ? t('noActiveExperiments') : t('pleaseSelectExperiment')}
                      </option>
                      {activeExperiments.map((exp) => (
                        <option key={exp.experiment_id} value={exp.experiment_id}>
                          {exp.name} ({exp.experiment_id.substring(0, 8)}...)
                        </option>
                      ))}
                    </select>

                    {experimentsLoadError && !loadingExperiments && (
                      <div style={{ marginTop: 6, fontSize: 11, color: '#ff4d4f', background: '#fff1f0', padding: '8px', borderRadius: 4 }}>
                        ❌ {t('loadFailed')} {experimentsLoadError}
                        <br />
                        <span style={{ color: '#8c8c8c' }}>{t('checkApiBaseUrl')}</span>
                      </div>
                    )}

                    {activeExperiments.length === 0 && !loadingExperiments && !experimentsLoadError && (
                      <div style={{ marginTop: 6, fontSize: 11, color: '#fa8c16' }}>
                        ⚠️ {t('noActiveExperimentsHint')}
                        <a href="/experiments" style={{ color: '#1677ff', textDecoration: 'underline' }}>
                          {t('experimentList')}
                        </a>
                        {t('activateExperimentHint')}
                      </div>
                    )}

                    {activeExperiments.length > 0 && !loadingExperiments && (
                      <div style={{ marginTop: 6, fontSize: 11, color: '#52c41a', background: '#f6ffed', padding: '8px', borderRadius: 4 }}>
                        ✅ {t('loadedFrom')} <code style={{ background: '#d9f7be', padding: '2px 6px', borderRadius: 3 }}>{apiBase}</code> {t('loaded')} {activeExperiments.length} {t('activeExperimentsCount')}
                      </div>
                    )}

                    {experimentId && (
                      <div style={{ marginTop: 6, fontSize: 11, color: '#52c41a' }}>
                        ✅ {t('selected')} <code style={{ background: '#f6ffed', padding: '2px 6px', borderRadius: 3 }}>{experimentId}</code>
                      </div>
                    )}
                  </div>

                  {/* User Key */}
                  <div className="compact-form-group">
                    <label className="compact-label">
                      <span className="label-text">{t('userKeyOptional')}</span>
                    </label>
                    <input
                      className="compact-input"
                      value={userKey}
                      onChange={(e) => setUserKey(e.target.value)}
                      placeholder="demo-user-123"
                    />
                    <div style={{ marginTop: 6, fontSize: 11, color: '#8c8c8c' }}>
                      {t('userKeyHint')}
                    </div>
                  </div>

                  {/* Random Assignment */}
                  <div className="compact-form-group">
                    <label className="compact-label" style={{ display: 'flex', alignItems: 'center' }}>
                      <input
                        type="checkbox"
                        checked={randomAssignment}
                        onChange={(e) => setRandomAssignment(e.target.checked)}
                        style={{ marginRight: 8 }}
                      />
                      <span className="label-text">{t('enableRandomAssignment')}</span>
                    </label>
                    <div style={{ marginTop: 6, fontSize: 11, color: '#8c8c8c' }}>
                      {t('randomAssignmentHint')}
                    </div>
                  </div>
                </div>

                {/* 操作按钮 */}
                <div className="compact-form-actions" style={{ marginTop: 20 }}>
                  <button className="compact-btn compact-btn-primary" onClick={loadPreview}>
                    <i className="fas fa-play"></i>
                    <span>{t('loadPreview')}</span>
                  </button>
                </div>
              </div>
            </div>

            {/* 预览和集成代码并排显示 */}
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '20px' }}>
              {/* 预览区域 */}
              <div className="compact-card">
                <div className="compact-card-header">
                  <h3 className="compact-card-title">{t('livePreview')}</h3>
                  <div className="compact-card-hint">
                    {previewKey > 0 ? t('previewLoaded') : t('clickToLoadPreview')}
                  </div>
                </div>

                <div className="compact-card-body">
                  {previewKey === 0 ? (
                    <div
                      style={{
                        minHeight: '300px',
                        display: 'flex',
                        flexDirection: 'column',
                        alignItems: 'center',
                        justifyContent: 'center',
                        border: '2px dashed #e8e8e8',
                        borderRadius: '8px',
                        color: '#8c8c8c',
                        padding: '40px',
                      }}
                    >
                      <i className="fas fa-play-circle" style={{ fontSize: '64px', marginBottom: '20px', color: '#d9d9d9' }}></i>
                      <h3 style={{ fontSize: '16px', marginBottom: '8px', color: '#595959' }}>{t('ready')}</h3>
                      <p style={{ fontSize: '14px', textAlign: 'center' }}>
                        {t('configureAndLoad')}<br />
                        {t('viewWidgetEffect')}
                      </p>
                    </div>
                  ) : (
                    <iframe
                      key={previewKey}
                      ref={iframeRef}
                      srcDoc={generatePreviewHTML()}
                      style={{
                        width: '100%',
                        height: '350px',
                        border: '1px solid #e8e8e8',
                        borderRadius: '6px',
                        background: 'white',
                      }}
                      sandbox="allow-scripts allow-same-origin"
                      title="Experiment Widget Preview"
                    />
                  )}
                </div>
              </div>

              {/* 集成代码 */}
              <div className="compact-card">
                <div className="compact-card-header">
                  <h3 className="compact-card-title">{t('integrationCode')}</h3>
                  <div className="compact-card-hint">{t('integrationCodeHint')}</div>
                </div>

                <div className="compact-card-body">
                <pre
                  style={{
                    background: '#f5f5f5',
                    padding: '16px',
                    borderRadius: '6px',
                    overflow: 'auto',
                    fontSize: '13px',
                    lineHeight: '1.6',
                    margin: 0,
                  }}
                >
                  <code>{`<!-- Experiment Widget SDK -->
<script
  src="${widgetSrc}"
  data-api-base="${apiBase}"
  data-experiment-id="${experimentId || 'YOUR_EXPERIMENT_ID'}"
  data-user-key="${userKey || 'OPTIONAL_USER_KEY'}"
  data-random-assignment="${randomAssignment}"
  async>
</script>`}</code>
                </pre>

                <div style={{ marginTop: 12, display: 'flex', justifyContent: 'flex-end' }}>
                  <button className="compact-btn compact-btn-primary" onClick={copyIntegrationCode}>
                    <i className="fas fa-copy"></i>
                    <span>{t('copyIntegrationCode')}</span>
                  </button>
                </div>

                <div style={{ marginTop: 16, padding: '12px', background: '#e6f4ff', borderRadius: '6px', fontSize: '13px' }}>
                  <div style={{ fontWeight: 600, marginBottom: 8, color: '#0958d9' }}>{t('integrationTips')}</div>
                  <ul style={{ margin: 0, paddingLeft: '20px', color: '#096dd9', lineHeight: '1.8' }}>
                    <li>{t('integrationTip1')}</li>
                    <li>{t('integrationTip2')}</li>
                    <li>{t('integrationTip3')}</li>
                    <li>{t('integrationTip4')}</li>
                  </ul>
                </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default PluginPreviewPage;
