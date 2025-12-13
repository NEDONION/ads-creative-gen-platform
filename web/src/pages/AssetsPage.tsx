import React, { useState, useEffect, useMemo } from 'react';
import { creativeAPI } from '../services/api';
import Sidebar from '../components/Sidebar';
import type { AssetData } from '../types';

const AssetsPage: React.FC = () => {
  const [assets, setAssets] = useState<AssetData[]>([]);
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const pageSize = 20;
  const CACHE_KEY = 'assets_page_cache_v1';
  const CACHE_TTL = 6 * 60 * 60 * 1000; // 6小时缓存

  useEffect(() => {
    const cacheRaw = sessionStorage.getItem(CACHE_KEY);
    if (cacheRaw) {
      try {
        const cached = JSON.parse(cacheRaw) as { ts: number; data: AssetData[]; total: number; total_pages: number; page: number };
        if (Date.now() - cached.ts < CACHE_TTL && cached.page === currentPage) {
          setAssets(cached.data);
          setTotal(cached.total);
          setTotalPages(cached.total_pages);
          return;
        }
      } catch {
        // ignore cache parse errors
      }
    }
    loadAssets();
  }, [currentPage]);

  const loadAssets = async (force?: boolean) => {
    if (!force) {
      // 优先使用缓存
      const cacheRaw = sessionStorage.getItem(CACHE_KEY);
      if (cacheRaw) {
        try {
          const cached = JSON.parse(cacheRaw) as { ts: number; data: AssetData[]; total: number; total_pages: number; page: number };
          if (Date.now() - cached.ts < CACHE_TTL && cached.page === currentPage) {
            setAssets(cached.data);
            setTotal(cached.total);
            setTotalPages(cached.total_pages);
            return;
          }
        } catch {
          // ignore
        }
      }
    }

    setLoading(true);
    setError(null);
    try {
      const response = await creativeAPI.listAssets({
        page: currentPage,
        page_size: pageSize,
      });

      if (response.code === 0 && response.data) {
        setAssets(response.data.assets || []);
        setTotal(response.data.total);
        setTotalPages(response.data.total_pages);
        // 写缓存
        sessionStorage.setItem(
          CACHE_KEY,
          JSON.stringify({
            ts: Date.now(),
            data: response.data.assets || [],
            total: response.data.total,
            total_pages: response.data.total_pages,
            page: currentPage,
          })
        );
      } else {
        setError(response.message || '获取素材列表失败');
      }
    } catch (err) {
      setError('加载素材失败: ' + (err as Error).message);
      console.error('Load assets error:', err);
    } finally {
      setLoading(false);
    }
  };

  const grouped = useMemo(() => {
    const groups: {
      key: string;
      productName: string;
      ctaText?: string;
      sellingPoints?: string[];
      assets: AssetData[];
    }[] = [];
    const map = new Map<string, number>();

    assets.forEach((asset) => {
      const key = asset.product_name || asset.title || '未命名创意';
      const idx = map.get(key);
      if (idx === undefined) {
        map.set(key, groups.length);
        groups.push({
          key,
          productName: key,
          ctaText: asset.cta_text,
          sellingPoints: asset.selling_points,
          assets: [asset],
        });
      } else {
        groups[idx].assets.push(asset);
        if (!groups[idx].ctaText && asset.cta_text) groups[idx].ctaText = asset.cta_text;
        if (!groups[idx].sellingPoints && asset.selling_points) groups[idx].sellingPoints = asset.selling_points;
      }
    });
    return groups;
  }, [assets]);

  return (
    <div className="app">
      <Sidebar />

      <div className="main-content">
        <div className="header">
          <h1 className="page-title">创意管理</h1>
          <div className="user-info">
            <div className="avatar">A</div>
            <span>管理员</span>
          </div>
        </div>

        <div className="content">
          <div className="compact-layout">
              <div className="compact-toolbar">
                <div className="compact-toolbar-left">
                  <div className="compact-stats-text">
                    共 <strong>{total}</strong> 个素材
                  </div>
                </div>
                <div className="compact-toolbar-right">
                  <button className="compact-btn compact-btn-outline compact-btn-sm" onClick={() => loadAssets(true)} disabled={loading}>
                    <i className="fas fa-sync"></i>
                    <span>刷新</span>
                  </button>
                </div>
              </div>

            {error && (
              <div className="compact-alert compact-alert-error">
                <i className="fas fa-exclamation-circle"></i>
                <span>{error}</span>
              </div>
            )}

            {loading ? (
              <div className="compact-loading">
                <div className="loading"></div>
                <div className="compact-loading-text">加载中...</div>
              </div>
            ) : assets.length === 0 ? (
              <div className="compact-empty">
                <i className="fas fa-images"></i>
                <div className="compact-empty-text">暂无素材</div>
                <div className="compact-empty-hint">生成创意后素材将显示在这里</div>
              </div>
            ) : (
              <>
                <div className="compact-card">
                  <div className="compact-card-header">
                    <h3 className="compact-card-title">按商品/创意分组</h3>
                    <div className="compact-card-hint">点击卡片可查看该商品的所有素材</div>
                  </div>
                  <div className="compact-card-body" style={{ display: 'grid', gap: 12 }}>
                    {grouped.map((group) => (
                      <div key={group.key} className="compact-card" style={{ border: '1px solid #f0f0f0' }}>
                        <div className="compact-card-header" style={{ borderBottom: '1px solid #f5f5f5' }}>
                          <div>
                            <h4 className="compact-card-title" style={{ margin: 0 }}>{group.productName}</h4>
                            {group.ctaText && <div style={{ fontSize: 12, color: '#8c8c8c' }}>CTA：{group.ctaText}</div>}
                            {group.sellingPoints && group.sellingPoints.length > 0 && (
                              <div style={{ fontSize: 12, color: '#8c8c8c' }}>卖点：{group.sellingPoints.join('、')}</div>
                            )}
                          </div>
                          <div style={{ fontSize: 12, color: '#8c8c8c' }}>{group.assets.length} 张素材</div>
                        </div>
                        <div className="compact-assets-grid">
                          {group.assets.map((asset) => (
                            <div key={asset.id} className="compact-asset-card">
                              <div className="compact-asset-image-wrapper">
                                <img
                                  src={asset.public_url || asset.image_url}
                                  alt={asset.id}
                                  className="compact-asset-image"
                                  onError={(e) => {
                                    const target = e.target as HTMLImageElement;
                                    target.src =
                                      'data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" width="240" height="180" viewBox="0 0 240 180"><rect width="240" height="180" fill="%23f5f5f5"/><text x="120" y="90" font-family="Arial" font-size="11" text-anchor="middle" fill="%23999">素材图片</text></svg>';
                                  }}
                                />
                              </div>
                              <div className="compact-asset-info">
                                <div className="compact-asset-meta">
                                  <span className="compact-asset-format">{asset.format}</span>
                                  <span className="compact-asset-size">
                                    {asset.width}×{asset.height}
                                  </span>
                                </div>
                                <div className="compact-asset-id" title={asset.id}>
                                  {asset.id.substring(0, 8)}...
                                </div>
                              </div>
                            </div>
                          ))}
                        </div>
                      </div>
                    ))}
                  </div>
                </div>

                {totalPages > 1 && (
                  <div className="compact-pagination">
                    <button
                      className="compact-page-btn"
                      disabled={currentPage <= 1}
                      onClick={() => setCurrentPage(currentPage - 1)}
                    >
                      <i className="fas fa-chevron-left"></i>
                    </button>

                    <div className="compact-page-numbers">
                      {Array.from({ length: Math.min(5, totalPages) }, (_, i) => {
                        const pageNum = i + Math.max(1, Math.min(currentPage - 2, totalPages - 4));
                        if (pageNum > totalPages) return null;
                        return (
                          <button
                            key={pageNum}
                            className={`compact-page-btn ${currentPage === pageNum ? 'active' : ''}`}
                            onClick={() => setCurrentPage(pageNum)}
                          >
                            {pageNum}
                          </button>
                        );
                      })}
                    </div>

                    <button
                      className="compact-page-btn"
                      disabled={currentPage >= totalPages}
                      onClick={() => setCurrentPage(currentPage + 1)}
                    >
                      <i className="fas fa-chevron-right"></i>
                    </button>

                    <div className="compact-page-info">
                      第 {currentPage} / {totalPages} 页
                    </div>
                  </div>
                )}
              </>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

export default AssetsPage;
