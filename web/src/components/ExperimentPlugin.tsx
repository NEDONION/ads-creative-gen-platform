import React, { useEffect, useMemo, useState } from 'react';
import { experimentAPI } from '../services/api';

// 本地定义，便于单文件复制使用
interface ExperimentAssignData {
  creative_id: number;
  asset_uuid?: string;
  task_id?: number;
  title?: string;
  product_name?: string;
  cta_text?: string;
  selling_points?: string[];
  image_url?: string;
}

interface ExperimentPluginProps {
  experimentId: string;
  userKey?: string;
  autoHit?: boolean; // 自动曝光埋点
  renderCreative?: (creativeId: number) => React.ReactNode;
  onAssigned?: (creativeId: number) => void;
  onHitTracked?: () => void;
  onClickTracked?: () => void;
}

/**
 * 轻量实验前端插件：
 * - 指定 experimentId，自动分流并返回 creativeId
 * - 可选自动曝光（autoHit），点击时调用 trackClick
 * - 可用 renderCreative 渲染创意内容
 */
const ExperimentPlugin: React.FC<ExperimentPluginProps> = ({
  experimentId,
  userKey,
  autoHit = true,
  renderCreative,
  onAssigned,
  onHitTracked,
  onClickTracked,
}) => {
  const [creativeId, setCreativeId] = useState<number | null>(null);
  const [creativeInfo, setCreativeInfo] = useState<ExperimentAssignData | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [hitDone, setHitDone] = useState(false);

  const canTrack = useMemo(() => !!creativeId, [creativeId]);

  // 分流
  useEffect(() => {
    const assign = async () => {
      if (!experimentId) return;
      setLoading(true);
      setError(null);
      setHitDone(false);
      try {
        const res = await experimentAPI.assign(experimentId, userKey);
        if (res.code === 0 && res.data) {
          setCreativeId(res.data.creative_id);
          setCreativeInfo(res.data);
          onAssigned?.(res.data.creative_id);
        } else {
          setError(res.message || '分流失败');
        }
      } catch (err) {
        setError((err as Error).message);
      } finally {
        setLoading(false);
      }
    };
    assign();
  }, [experimentId, userKey, onAssigned]);

  // 自动曝光
  useEffect(() => {
    if (!autoHit || !creativeId || hitDone) return;
    const doHit = async () => {
      try {
        await experimentAPI.hit(experimentId, creativeId);
        setHitDone(true);
        onHitTracked?.();
      } catch (err) {
        console.error('hit failed:', err);
      }
    };
    doHit();
  }, [autoHit, creativeId, experimentId, hitDone, onHitTracked]);

  const trackClick = async () => {
    if (!canTrack) return;
    try {
      await experimentAPI.click(experimentId, creativeId as number);
      onClickTracked?.();
    } catch (err) {
      console.error('click failed:', err);
    }
  };

  return (
    <div style={{ border: '1px dashed #ddd', padding: 12, borderRadius: 8, fontSize: 14 }}>
      <div style={{ marginBottom: 8, color: '#555' }}>
        实验ID: <code>{experimentId}</code>
      </div>
      {loading && <div>正在分流...</div>}
      {error && <div style={{ color: '#c00' }}>分流失败：{error}</div>}
      {creativeId && (
        <div>
          <div style={{ marginBottom: 8 }}>
            命中创意ID: <strong>{creativeId}</strong>
          </div>
          {creativeInfo && (
            <div style={{ display: 'flex', gap: 12, alignItems: 'center', marginBottom: 8 }}>
              {creativeInfo.image_url ? (
                <img
                  src={creativeInfo.image_url}
                  alt="创意缩略图"
                  style={{ width: 96, height: 96, objectFit: 'cover', borderRadius: 8, border: '1px solid #eee' }}
                />
              ) : (
                <div style={{ width: 96, height: 96, borderRadius: 8, border: '1px dashed #ccc', display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#999', fontSize: 12 }}>
                  无图
                </div>
              )}
              <div style={{ flex: 1 }}>
                <div style={{ fontWeight: 700 }}>{creativeInfo.title || creativeInfo.product_name || '创意'}</div>
                <div style={{ fontSize: 13, color: '#555', marginTop: 4 }}>
                  {creativeInfo.product_name || '未设置商品名'}
                </div>
                {creativeInfo.selling_points && creativeInfo.selling_points.length > 0 && (
                  <div style={{ fontSize: 12, color: '#666', marginTop: 4, lineHeight: 1.4 }}>
                    卖点：{creativeInfo.selling_points.slice(0, 2).join(' / ')}
                  </div>
                )}
                {creativeInfo.cta_text && (
                  <div style={{ fontSize: 12, color: '#111', marginTop: 4, fontWeight: 600 }}>
                    CTA：{creativeInfo.cta_text}
                  </div>
                )}
              </div>
            </div>
          )}
          {renderCreative ? (
            <div>{renderCreative(creativeId)}</div>
          ) : (
            <div style={{ color: '#666' }}>在 renderCreative 中渲染实际素材。</div>
          )}
          <div style={{ marginTop: 8, display: 'flex', gap: 8 }}>
            <button
              className="compact-btn compact-btn-outline compact-btn-xs"
              onClick={() => experimentAPI.hit(experimentId, creativeId)}
              disabled={!canTrack}
            >
              手动曝光
            </button>
            <button className="compact-btn compact-btn-primary compact-btn-xs" onClick={trackClick} disabled={!canTrack}>
              记录点击
            </button>
          </div>
        </div>
      )}
    </div>
  );
};

export default ExperimentPlugin;
