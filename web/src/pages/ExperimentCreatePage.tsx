import React, { useEffect, useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import Sidebar from '../components/Sidebar';
import { experimentAPI, creativeAPI } from '../services/api';
import type { ExperimentVariantInput } from '../types';
import LanguageSwitch from '../components/LanguageSwitch';
import { useI18n } from '../i18n';

const ExperimentCreatePage: React.FC = () => {
  const navigate = useNavigate();
  const { t } = useI18n();
  const [name, setName] = useState('');
  const [productName, setProductName] = useState('');
  const [variants, setVariants] = useState<ExperimentVariantInput[]>([{ creative_id: '', weight: 0.5 }, { creative_id: '', weight: 0.5 }]);
  const [creating, setCreating] = useState(false);
  const [message, setMessage] = useState<string | null>(null);
  const [productOptions, setProductOptions] = useState<string[]>([]);
  const [creativeOptions, setCreativeOptions] = useState<{ id: string; label: string; thumb?: string; product_name?: string; cta_text?: string; selling_points?: string[]; title?: string }[]>([]);

  useEffect(() => {
    loadOptions();
  }, []);

  const loadOptions = async () => {
    try {
      const [tasksRes, assetsRes] = await Promise.all([
        creativeAPI.listTasks({ page: 1, page_size: 100 }),
        creativeAPI.listAssets({ page: 1, page_size: 100 }),
      ]);

      if (tasksRes.code === 0 && tasksRes.data) {
        const names = (tasksRes.data.tasks || [])
          .map((t) => t.product_name)
          .filter((n): n is string => Boolean(n));
        const uniqueNames = Array.from(new Set(names));
        setProductOptions(uniqueNames);
        if (!productName && uniqueNames.length > 0) {
          setProductName(uniqueNames[0]);
        }
      }

      if (assetsRes.code === 0 && assetsRes.data) {
        const taskProductMap: Record<string, string | undefined> = {};
        (tasksRes.data?.tasks || []).forEach((t) => {
          if (t.id && t.product_name) {
            taskProductMap[t.id] = t.product_name;
          }
        });

        const options = (assetsRes.data.assets || [])
          .map((asset) => {
            const resolvedProduct = asset.product_name || taskProductMap[String(asset.task_id)];
            const label = asset.title || resolvedProduct || t('creativeId');
            return {
              id: asset.id,
              label: `${label} (${asset.id})`,
              thumb: asset.image_url || asset.public_url,
              product_name: resolvedProduct,
              cta_text: asset.cta_text,
              selling_points: asset.selling_points,
              title: asset.title,
            };
          })
          .filter((v) => Boolean(v && v.id));
        setCreativeOptions(options);
      }
    } catch (err) {
      console.error(err);
      setMessage(t('activityError'));
    }
  };

  const updateVariant = (idx: number, field: keyof ExperimentVariantInput, value: number | string | string[]) => {
    setVariants((prev) => {
      const copy = [...prev];
      copy[idx] = { ...copy[idx], [field]: value } as ExperimentVariantInput;
      return copy;
    });
  };

  const addVariant = () => {
    setVariants((prev) => [...prev, { creative_id: '', weight: 0.1, cta_text: '', selling_points: [] }]);
  };

  const removeVariant = (idx: number) => {
    setVariants((prev) => prev.filter((_, i) => i !== idx));
  };

  const findCreativeOption = (id: string | number) => creativeOptions.find((opt) => String(opt.id) === String(id));

  const getCreativeLabel = (id: string | number) => {
    const opt = findCreativeOption(id);
    if (opt) return opt.label;
    return String(id);
  };

  const handleCreate = async () => {
    if (!name.trim()) {
      setMessage(t('fillName'));
      return;
    }
    if (variants.some((v) => !v.creative_id || v.weight <= 0)) {
      setMessage(t('fillVariant'));
      return;
    }
    setCreating(true);
    setMessage(null);
    try {
      const res = await experimentAPI.create({ name: name.trim(), product_name: productName.trim(), variants });
      if (res.code === 0 && res.data) {
        setMessage(`${t('createSuccess')} ID: ${res.data.experiment_id}`);
        navigate('/experiments');
      } else {
        setMessage(res.message || t('createFail'));
      }
    } catch (err) {
      setMessage(t('createFail') + ': ' + (err as Error).message);
    } finally {
      setCreating(false);
    }
  };

  const filteredCreativeOptions = productName
    ? creativeOptions.filter((opt) => (opt.product_name || '').trim() === productName.trim())
    : creativeOptions;

  const summary = useMemo(() => {
    const selectedCreatives = variants.map((v) => v.creative_id).filter(Boolean) as (string | number)[];
    return {
      productName: productName || t('productLabel'),
      variantCount: variants.length,
      creatives: selectedCreatives,
    };
  }, [variants, productName, t]);

  return (
    <div className="app">
      <Sidebar />

      <div className="main-content">
        <div className="header">
          <h1 className="page-title">{t('headerExperimentNew')}</h1>
          <div className="user-info">
            <LanguageSwitch />
            <div className="avatar">A</div>
            <span>{t('admin')}</span>
          </div>
        </div>

        <div className="content">
          <div className="compact-layout">
            {message && (
              <div className="compact-alert compact-alert-info">
                <i className="fas fa-info-circle"></i>
                <span>{message}</span>
              </div>
            )}

            <div className="compact-card">
              <div className="compact-card-header">
                <h3 className="compact-card-title">{t('createExperiment')}</h3>
                <div className="compact-card-hint">{t('createExperimentHint')}</div>
              </div>
              <div className="compact-card-body">
                <div className="compact-form-grid">
                  <div className="compact-form-group">
                    <label className="compact-label">
                      <span className="label-text">{t('nameLabel')}</span>
                      <span className="label-required">*</span>
                    </label>
                    <input className="compact-input" value={name} onChange={(e) => setName(e.target.value)} />
                  </div>
                  <div className="compact-form-group">
                    <label className="compact-label">
                      <span className="label-text">{t('productLabel')}</span>
                    </label>
                    <select className="compact-input" value={productName} onChange={(e) => setProductName(e.target.value)}>
                      {productOptions.length === 0 && <option value="">{t('noProduct')}</option>}
                      {productOptions.map((p) => (
                        <option key={p} value={p}>
                          {p}
                        </option>
                      ))}
                    </select>
                  </div>
                </div>

                <div className="compact-section-title" style={{ marginTop: 8 }}>{t('variantConfigLabel')}</div>
                <div className="compact-form-grid">
                  {variants.map((v, idx) => {
                    const creative = findCreativeOption(v.creative_id);
                    return (
                      <div key={idx} className="compact-card" style={{ padding: 10, border: '1px solid #f0f0f0' }}>
                        <div className="compact-form-group">
                          <label className="compact-label">{t('creativeId')}</label>
                          <select
                            className="compact-input"
                            value={v.creative_id || ''}
                            onChange={(e) => updateVariant(idx, 'creative_id', e.target.value)}
                          >
                            <option value="">{t('selectCreative')}</option>
                            {filteredCreativeOptions.map((opt) => (
                              <option key={opt.id} value={opt.id}>
                                {opt.label}
                              </option>
                            ))}
                          </select>
                          {v.creative_id && (
                            <div className="compact-thumb" style={{ marginTop: 6, display: 'flex', alignItems: 'center', gap: 8 }}>
                              {creative?.thumb ? (
                                <img
                                  src={creative.thumb}
                                  alt={t('creativeId')}
                                  style={{ width: 80, height: 80, objectFit: 'cover', borderRadius: 6, border: '1px solid #eee' }}
                                />
                              ) : (
                                <div style={{ width: 80, height: 80, borderRadius: 6, border: '1px dashed #ccc', display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#999', fontSize: 12 }}>
                                  {t('noThumb')}
                                </div>
                              )}
                              <div style={{ fontSize: 12, color: '#555' }}>
                                {getCreativeLabel(v.creative_id)}
                              </div>
                            </div>
                          )}
                        </div>
                        <div className="compact-form-group">
                          <label className="compact-label">{t('weight')} (0-1)</label>
                          <input
                            className="compact-input"
                            type="number"
                            step="0.1"
                            value={v.weight}
                            onChange={(e) => updateVariant(idx, 'weight', parseFloat(e.target.value) || 0)}
                          />
                        </div>
                        <div className="compact-form-group">
                          <label className="compact-label">{t('ctaOverride')}</label>
                          {(() => {
                            const creativeCTA = creative?.cta_text || '';
                            const selectedCTA = (v as any).cta_text || '';
                            const options = [{ value: '', label: t('useDefault') }];
                            if (creativeCTA) {
                              options.push({ value: creativeCTA, label: `${t('creativeCTA')}${creativeCTA}` });
                            }
                            if (selectedCTA && selectedCTA !== creativeCTA) {
                              options.push({ value: selectedCTA, label: `${t('currentOverride')}${selectedCTA}` });
                            }
                            return (
                              <>
                                <select
                                  className="compact-input"
                                  value={selectedCTA}
                                  onChange={(e) => updateVariant(idx, 'cta_text' as keyof ExperimentVariantInput, e.target.value)}
                                >
                                  {options.map((opt) => (
                                    <option key={opt.value || 'default'} value={opt.value}>
                                      {opt.label}
                                    </option>
                                  ))}
                                </select>
                                <input
                                  className="compact-input"
                                  style={{ marginTop: 6 }}
                                  type="text"
                                  placeholder={t('ctaPlaceholder')}
                                  value={selectedCTA}
                                  onChange={(e) => updateVariant(idx, 'cta_text' as keyof ExperimentVariantInput, e.target.value)}
                                />
                              </>
                            );
                          })()}
                        </div>
                        <div className="compact-form-group">
                          <label className="compact-label">{t('spOverride')}</label>
                          {(() => {
                            const spOptions = Array.isArray(creative?.selling_points) ? creative.selling_points : [];
                            const selectedSP: string[] = Array.isArray((v as any).selling_points) ? ((v as any).selling_points as any) : [];
                            const toggleSP = (sp: string) => {
                              const next = selectedSP.includes(sp) ? selectedSP.filter((s: string) => s !== sp) : [...selectedSP, sp];
                              updateVariant(idx, 'selling_points' as keyof ExperimentVariantInput, next);
                            };
                            return spOptions.length > 0 ? (
                              <div className="option-grid">
                                {spOptions.map((sp) => (
                                  <label key={sp} className={`checkbox-option ${selectedSP.includes(sp) ? 'active' : ''}`}>
                                    <input type="checkbox" checked={selectedSP.includes(sp)} onChange={() => toggleSP(sp)} />
                                    <span className="checkbox-label">{sp}</span>
                                  </label>
                                ))}
                              </div>
                            ) : (
                              <div style={{ color: '#999', fontSize: 12 }}>{t('pendingGen')}</div>
                            );
                          })()}
                        </div>
                        {v.creative_id && (
                          <div className="compact-form-group">
                            <label className="compact-label">{t('copywriting')}</label>
                            <div style={{ background: '#fafafa', border: '1px solid #eee', borderRadius: 6, padding: 8, fontSize: 12, color: '#555' }}>
                              {(() => {
                                return creative?.cta_text ? (
                                  <div style={{ marginBottom: 4 }}>
                                    <strong>CTA：</strong>
                                    <span>{creative.cta_text}</span>
                                  </div>
                                ) : (
                                  <div style={{ marginBottom: 4, color: '#999' }}>{t('pendingGen')}</div>
                                );
                              })()}
                              {(() => {
                                return creative?.selling_points && creative.selling_points.length > 0 ? (
                                  <div>
                                    <strong>{t('sellingPointsLabel')}：</strong>
                                    <span>{creative.selling_points.slice(0, 3).join(' / ')}</span>
                                  </div>
                                ) : (
                                  <div style={{ color: '#999' }}>{t('pendingGen')}</div>
                                );
                              })()}
                            </div>
                          </div>
                        )}
                        {variants.length > 2 && (
                          <button className="compact-btn compact-btn-text compact-btn-xs" onClick={() => removeVariant(idx)}>
                            {t('delete')}
                          </button>
                        )}
                      </div>
                    );
                  })}
                </div>
                <div className="compact-form-actions" style={{ marginTop: 8, display: 'flex', gap: 10 }}>
                  <button className="compact-btn compact-btn-outline compact-btn-sm" onClick={addVariant}>
                    <i className="fas fa-plus"></i>
                    <span>{t('addVariantBtn')}</span>
                  </button>
                  <button className="compact-btn compact-btn-primary" onClick={handleCreate} disabled={creating}>
                    <i className="fas fa-save"></i>
                    <span>{creating ? t('submitting') : t('submitCreate')}</span>
                  </button>
                  <button className="compact-btn compact-btn-text compact-btn-sm" onClick={() => navigate('/experiments')}>
                    {t('backToList')}
                  </button>
                </div>
              </div>
            </div>

            <div className="compact-card">
              <div className="compact-card-header">
                <h3 className="compact-card-title">{t('currentConfig')}</h3>
                <div className="compact-card-hint">{t('variantConfigLabel')}</div>
              </div>
              <div className="compact-card-body" style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit,minmax(180px,1fr))', gap: 10 }}>
                <div className="meta-block">
                  <div className="meta-label">{t('productLabel')}</div>
                  <div className="meta-value">{summary.productName || t('notFilled')}</div>
                </div>
                <div className="meta-block">
                  <div className="meta-label">{t('variantTitle')}</div>
                  <div className="meta-value">{summary.variantCount}</div>
                </div>
                <div className="meta-block" style={{ gridColumn: '1 / -1' }}>
                  <div className="meta-label">{t('creativeId')}</div>
                  <div className="meta-value" style={{ display: 'flex', gap: 6, flexWrap: 'wrap' }}>
                      {(summary.creatives as (string | number)[]).length === 0 ? (
                        <span style={{ color: '#999' }}>{t('selectCreative')}</span>
                      ) : (
                        (summary.creatives as (string | number)[]).map((c, i) => (
                          <code key={`${c}-${i}`} className="compact-code">
                            {c}
                          </code>
                        ))
                      )}
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

export default ExperimentCreatePage;
