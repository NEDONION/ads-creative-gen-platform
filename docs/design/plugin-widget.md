# 可收起浮窗广告插件设计与使用文档

## 文档版本
- 版本：v1.2
- 日期：2025-12-xx
- 状态：实现版（可直接在站点/React 中使用）

---

## 1. 概述
在任意个人网站以脚本方式嵌入一个“可收起的浮窗广告”插件，展示命中的实验创意，自动分流、曝光、点击上报，样式轻量不污染宿主页面。

---

## 2. 功能与体验
- **浮窗展示**：可收起/展开，默认右下角；显示创意缩略/文案。
- **分流支持**：必填 `experimentId`，按用户键分桶返回 `creativeId`。
- **埋点上报**：自动曝光一次；点击上报；可自定义点击跳转。
- **低侵入**：纯前端脚本，类名前缀隔离；支持移动端宽度自适应。

---

## 3. 嵌入方式（纯 HTML，不依赖框架）
```html
<div id="exp-widget"></div>
<style>/* 样式见附录 A，可直接粘贴 */</style>

<script type="module">
const API_BASE = 'http://localhost:4000/api/v1'; // 改成你的后端
const experimentId = 'your-experiment-id';       // 填你的实验 ID
const userKey = '';                              // 可选，用于稳定分桶

async function api(path, opts={}) {
  const res = await fetch(API_BASE + path, {
    headers: { 'Content-Type': 'application/json' },
    ...opts,
    body: opts.body ? JSON.stringify(opts.body) : undefined,
  });
  return res.json();
}

const root = document.getElementById('exp-widget');
root.innerHTML = `
  <div class="ad">
    <div class="hdr"><span>赞助内容</span><div class="pill" id="exp-toggle">收起</div></div>
    <div class="body" id="exp-body">
      <div id="exp-status" style="font-size:13px;color:#cbd5e1;">分流中...</div>
      <div class="card" id="exp-card" style="display:none">
        <div class="th" id="exp-thumb"></div>
        <div class="info">
          <div class="title" id="exp-title"></div>
          <div class="meta">点击查看详情</div>
        </div>
      </div>
    </div>
  </div>
`;

const bodyEl = document.getElementById('exp-body');
const statusEl = document.getElementById('exp-status');
const cardEl = document.getElementById('exp-card');
const titleEl = document.getElementById('exp-title');
let creativeId = null;
let open = true;

document.querySelector('#exp-toggle').onclick = () => {
  open = !open;
  bodyEl.style.display = open ? 'block' : 'none';
  document.querySelector('#exp-toggle').innerText = open ? '收起' : '展开';
};

async function assign() {
  statusEl.innerText = '分流中...';
  cardEl.style.display = 'none';
  try {
    const res = await api(`/experiments/${experimentId}/assign?user_key=${encodeURIComponent(userKey||'')}`);
    if (res.code === 0 && res.data) {
      creativeId = res.data.creative_id;
      statusEl.style.display = 'none';
      titleEl.innerText = `命中创意 #${creativeId}`;
      cardEl.style.display = 'flex';
      // 自动曝光
      api(`/experiments/${experimentId}/hit`, { method:'POST', body:{ creative_id: creativeId } });
    } else {
      statusEl.innerText = res.message || '分流失败';
    }
  } catch (err) {
    statusEl.innerText = '分流异常：' + err.message;
  }
}

cardEl.onclick = () => {
  if (!creativeId) return;
  api(`/experiments/${experimentId}/click`, { method:'POST', body:{ creative_id: creativeId } });
  // TODO: 在这里做跳转/弹窗
};

assign();
</script>
```
> 若要显示真实缩略图：在 `assign()` 后调用你的素材接口，给 `#exp-thumb` 设置 `style.backgroundImage = 'url(...)'`。

---

## 4. React/TS 项目用法（推荐）
1) 复制 `web/src/components/ExperimentPlugin.tsx` 到你的项目（调整 `experimentAPI` 引用）；附加前文的 CSS。  
2) 页面使用：
```tsx
<ExperimentPlugin
  experimentId="your-experiment-id"
  userKey="user-123"           // 可选
  renderCreative={(id) => <div>命中创意 {id}</div>}
  autoHit
/>
```
3) 点击跳转：在组件的点击处理里添加你的跳转逻辑，保留 `experimentAPI.click` 调用。

---

## 5. 配置项（React 版）
```ts
type Props = {
  experimentId: string; // 必填
  userKey?: string;     // 可选，分桶稳定
  autoHit?: boolean;    // 默认 true，自动曝光
  renderCreative?: (creativeId: number) => React.ReactNode; // 自定义展示
  onAssigned?: (creativeId: number) => void;
  onHitTracked?: () => void;
  onClickTracked?: () => void;
};
```

---

## 6. 数据流与接口
1) 分流：`GET /experiments/:id/assign?user_key=...` → `{ creative_id }`  
2) 曝光：`POST /experiments/:id/hit { creative_id }`（自动一次，可手动）  
3) 点击：`POST /experiments/:id/click { creative_id }`（点击时触发）  
4) 素材：需另行请求创意/素材接口拿图片或落地页 URL（插件未自带）。

---

## 7. 设计要点（已实现的默认样式）
- 右下角浮窗、圆角+玻璃拟态背景；收起/展开切换。
- 卡片 hover 轻微提升 + 阴影；移动端宽度 90vw 自适应。
- 文案/按钮可改，类名前缀 `ad-widget` 防冲突。

---

## 8. 常见问题
- **CORS 拦截**：确认后端允许前端站点域名；或在 dev 用代理。  
- **分流失败**：检查 `experimentId` 是否存在且 active。  
- **曝光不上报**：检查 `/hit` 返回；必要时在分流成功后手动调用。  
- **缩略图未显示**：未内置素材获取，请按 creativeId 拉取素材后设置背景图。

---

## 附录 A：CSS（与 HTML 片段配套，可直接粘贴）
```css
#exp-widget .ad{position:fixed;right:20px;bottom:20px;width:280px;background:#0f172a;color:#e2e8f0;border-radius:16px;
box-shadow:0 16px 40px rgba(0,0,0,0.25);border:1px solid rgba(255,255,255,0.08);font-family:'Inter','PingFang SC',system-ui;overflow:hidden;z-index:9999;}
#exp-widget .hdr{padding:12px 14px;display:flex;align-items:center;justify-content:space-between;cursor:pointer;
background:linear-gradient(120deg,#1d2b64,#1d976c);font-weight:600;letter-spacing:.2px;}
#exp-widget .pill{background:rgba(255,255,255,0.15);border-radius:999px;padding:4px 10px;font-size:12px;}
#exp-widget .body{padding:12px;} #exp-widget .err{color:#fca5a5;font-size:13px;}
#exp-widget .card{display:flex;gap:12px;padding:10px;border-radius:12px;background:rgba(255,255,255,0.04);
border:1px solid rgba(255,255,255,0.08);cursor:pointer;transition:.2s;}
#exp-widget .card:hover{transform:translateY(-2px);box-shadow:0 10px 30px rgba(0,0,0,0.25);border-color:rgba(255,255,255,0.16);}
#exp-widget .th{width:64px;height:64px;border-radius:10px;background:linear-gradient(135deg,#60a5fa,#7c3aed);}
#exp-widget .info{display:flex;flex-direction:column;justify-content:center;gap:4px;}
#exp-widget .title{font-weight:700;font-size:14px;} #exp-widget .meta{font-size:12px;color:#94a3b8;}
@media(max-width:640px){#exp-widget .ad{width:90vw;right:5vw;bottom:12px;}}
```
