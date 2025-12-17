import React, { useState, useEffect, useMemo } from 'react';
import { creativeAPI } from '../services/api';
import Sidebar from '../components/Sidebar';
import type { AssetData } from '../types';
import Header from '../components/Header';
import { useI18n } from '../i18n';

const AssetsPage: React.FC = () => {
  const { t, lang } = useI18n();
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
  }, [currentPage, t]);

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
        setError(response.message || t('loadAssetsFailed'));
      }
    } catch (err) {
      setError(t('loadAssetsError').replace('{msg}', (err as Error).message));
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
      const key = asset.product_name || asset.title || t('unnamedCreative');
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
  }, [assets, lang, t]);


  return (
    <div className="app">
      <Sidebar />

      <div className="main-content">
        <Header title={t('headerAssets')} />

        <div className="content">
          <div className="compact-layout">
              <div className="compact-toolbar">
                <div className="compact-toolbar-left">
                  <div className="compact-stats-text">
                    {t('totalAssetsText').replace('{n}', String(total))}
                  </div>
                </div>
                <div className="compact-toolbar-right">
                  <button className="compact-btn compact-btn-outline compact-btn-sm" onClick={() => loadAssets(true)} disabled={loading}>
                    <i className="fas fa-sync"></i>
                    <span>{t('refresh')}</span>
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
                <div className="compact-loading-text">{t('loading')}</div>
              </div>
            ) : assets.length === 0 ? (
              <div className="compact-empty">
                <i className="fas fa-images"></i>
                <div className="compact-empty-text">{t('emptyAssets')}</div>
                <div className="compact-empty-hint">{t('emptyAssetsHint')}</div>
              </div>
            ) : (
              <>
                <div className="compact-card">
                  <div className="compact-card-header">
                    <h3 className="compact-card-title">{t('groupByProduct')}</h3>
                    <div className="compact-card-hint">{t('groupByHint')}</div>
                  </div>
                  <div className="compact-card-body assets-group-grid">
                    {grouped.map((group) => (
                      <div key={group.key} className="assets-group-card">
                        <div className="group-header">
                          <div>
                            <div className="group-title">{group.productName}</div>
                            {group.ctaText && <div className="group-sub">{t('cta')}: {group.ctaText}</div>}
                            {group.sellingPoints && group.sellingPoints.length > 0 && (
                              <div className="group-sub">
                                {t('sellingPointsLabel')}: {group.sellingPoints.join(lang === 'zh' ? '、' : ', ')}
                              </div>
                            )}
                          </div>
                          <div className="group-badge">{t('assetCountBadge').replace('{n}', String(group.assets.length))}</div>
                        </div>
                        <div className="compact-assets-vertical">
                          {group.assets.map((asset) => (
                            <div key={asset.id} className="vertical-asset-card">
                              <div className="vertical-meta">
                                <div className="meta-left">
                                  <span className="compact-asset-format">{asset.format}</span>
                                  <span className="compact-asset-size">{asset.width}×{asset.height}</span>
                                  <code className="compact-asset-id">{asset.id.substring(0, 8)}...</code>
                                </div>
                              </div>
                              <div className="vertical-body">
                                <div className="compact-asset-image-wrapper tall-vertical">
                                  <img
                                    src={asset.public_url || asset.image_url}
                                    alt={asset.id}
                                    className="compact-asset-image"
                                    onError={(e) => {
                                      const target = e.target as HTMLImageElement;
                                      target.src =
                                        `data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" width="280" height="320" viewBox="0 0 280 320"><rect width="280" height="320" fill="%23f5f5f5"/><text x="140" y="160" font-family="Arial" font-size="11" text-anchor="middle" fill="%23999">${t('assetPlaceholder')}</text></svg>`;
                                    }}
                                  />
                                </div>
                                <div className="compact-asset-info long vertical-info">
                                  {asset.title && (
                                    <div className="info-row">
                                      <span className="label">{t('titleLabel')}</span>
                                      <span className="value">{asset.title}</span>
                                    </div>
                                  )}
                                  {asset.cta_text && (
                                    <div className="info-row">
                                      <span className="label">CTA</span>
                                      <span className="value">{asset.cta_text}</span>
                                    </div>
                                  )}
                                  {asset.selling_points && asset.selling_points.length > 0 && (
                                    <div className="info-row">
                                      <span className="label">{t('sellingPointsLabel')}</span>
                                      <span className="value">{asset.selling_points.join(lang === 'zh' ? '、' : ' / ')}</span>
                                    </div>
                                  )}
                                  {asset.style && (
                                    <div className="info-row">
                                      <span className="label">{t('style')}</span>
                                      <span className="value">
                                        <span className="style-chip">{asset.style}</span>
                                      </span>
                                    </div>
                                  )}
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
                      {t('pageInfo').replace('{current}', String(currentPage)).replace('{total}', String(totalPages))}
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
