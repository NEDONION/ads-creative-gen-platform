import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import Sidebar from '../components/Sidebar';
import { experimentAPI, creativeAPI } from '../services/api';
import type { ExperimentVariantInput } from '../types';

const ExperimentCreatePage: React.FC = () => {
  const navigate = useNavigate();
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
            const label = asset.title || resolvedProduct || '创意';
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
      setMessage('加载商品/创意选项失败');
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

  const getCreativeLabel = (id: number | string) => {
    const opt = creativeOptions.find((o) => o.id === String(id));
    if (opt) return opt.label;
    return id;
  };

  const handleCreate = async () => {
    if (!name.trim()) {
      setMessage('请填写实验名称');
      return;
    }
    if (variants.some((v) => !v.creative_id || v.weight <= 0)) {
      setMessage('请填写有效的创意ID和权重');
      return;
    }
    setCreating(true);
    setMessage(null);
    try {
      const res = await experimentAPI.create({ name: name.trim(), product_name: productName.trim(), variants });
      if (res.code === 0 && res.data) {
        setMessage(`创建成功，ID: ${res.data.experiment_id}`);
        navigate('/experiments');
      } else {
        setMessage(res.message || '创建失败');
      }
    } catch (err) {
      setMessage('创建失败: ' + (err as Error).message);
    } finally {
      setCreating(false);
    }
  };

  const filteredCreativeOptions = productName
    ? creativeOptions.filter((opt) => (opt.product_name || '').trim() === productName.trim())
    : creativeOptions;

  return (
    <div className="app">
      <Sidebar />

      <div className="main-content">
        <div className="header">
          <h1 className="page-title">新建实验</h1>
          <div className="user-info">
            <div className="avatar">A</div>
            <span>管理员</span>
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
                <h3 className="compact-card-title">创建实验</h3>
                <div className="compact-card-hint">配置变体和权重，创建后可在实验列表查看</div>
              </div>
              <div className="compact-card-body">
                <div className="compact-form-grid">
                  <div className="compact-form-group">
                    <label className="compact-label">
                      <span className="label-text">实验名称</span>
                      <span className="label-required">*</span>
                    </label>
                    <input className="compact-input" value={name} onChange={(e) => setName(e.target.value)} />
                  </div>
                  <div className="compact-form-group">
                    <label className="compact-label">
                      <span className="label-text">商品名</span>
                    </label>
                    <select className="compact-input" value={productName} onChange={(e) => setProductName(e.target.value)}>
                      {productOptions.length === 0 && <option value="">暂无商品，请先创建任务</option>}
                      {productOptions.map((p) => (
                        <option key={p} value={p}>
                          {p}
                        </option>
                      ))}
                    </select>
                  </div>
                </div>

                <div className="compact-section-title" style={{ marginTop: 8 }}>变体配置</div>
                <div className="compact-form-grid">
                  {variants.map((v, idx) => (
                    <div key={idx} className="compact-card" style={{ padding: 10, border: '1px solid #f0f0f0' }}>
                      <div className="compact-form-group">
                        <label className="compact-label">创意ID</label>
                        <select
                          className="compact-input"
                          value={v.creative_id || ''}
                          onChange={(e) => updateVariant(idx, 'creative_id', e.target.value)}
                        >
                          <option value="">请选择创意</option>
                          {filteredCreativeOptions.map((opt) => (
                            <option key={opt.id} value={opt.id}>
                              {opt.label}
                            </option>
                          ))}
                        </select>
                        {v.creative_id && (
                          <div className="compact-thumb" style={{ marginTop: 6, display: 'flex', alignItems: 'center', gap: 8 }}>
                            {creativeOptions.find((opt) => opt.id === v.creative_id)?.thumb ? (
                              <img
                                src={creativeOptions.find((opt) => opt.id === v.creative_id)!.thumb}
                                alt="预览"
                                style={{ width: 80, height: 80, objectFit: 'cover', borderRadius: 6, border: '1px solid #eee' }}
                              />
                            ) : (
                              <div style={{ width: 80, height: 80, borderRadius: 6, border: '1px dashed #ccc', display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#999', fontSize: 12 }}>
                                无缩略图
                              </div>
                            )}
                            <div style={{ fontSize: 12, color: '#555' }}>
                              {getCreativeLabel(v.creative_id)}
                            </div>
                          </div>
                        )}
                      </div>
                      <div className="compact-form-group">
                        <label className="compact-label">权重 (0-1)</label>
                        <input
                          className="compact-input"
                          type="number"
                          step="0.1"
                          value={v.weight}
                          onChange={(e) => updateVariant(idx, 'weight', parseFloat(e.target.value) || 0)}
                        />
                      </div>
                      <div className="compact-form-group">
                        <label className="compact-label">CTA（选择或覆盖）</label>
                        {(() => {
                          const info = (creativeOptions.find((opt) => opt.id === v.creative_id) || {}) as { cta_text?: string };
                          const creativeCTA = info.cta_text || '';
                          const selectedCTA = (v as any).cta_text || '';
                          const options = [{ value: '', label: '使用创意默认' }];
                          if (creativeCTA) {
                            options.push({ value: creativeCTA, label: `创意：${creativeCTA}` });
                          }
                          if (selectedCTA && selectedCTA !== creativeCTA) {
                            options.push({ value: selectedCTA, label: `当前覆盖：${selectedCTA}` });
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
                                placeholder="手动输入覆盖 CTA"
                                value={selectedCTA}
                                onChange={(e) => updateVariant(idx, 'cta_text' as keyof ExperimentVariantInput, e.target.value)}
                              />
                            </>
                          );
                        })()}
                      </div>
                      <div className="compact-form-group">
                        <label className="compact-label">卖点选择（可多选）</label>
                        {(() => {
                          const info = (creativeOptions.find((opt) => opt.id === v.creative_id) || {}) as { selling_points?: string[] };
                          const spOptions = Array.isArray(info.selling_points) ? info.selling_points : [];
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
                            <div style={{ color: '#999', fontSize: 12 }}>该创意暂无卖点，可在创意生成时补充</div>
                          );
                        })()}
                      </div>
                      {v.creative_id && (
                        <div className="compact-form-group">
                          <label className="compact-label">当前创意文案（回填自生成时）</label>
                          <div style={{ background: '#fafafa', border: '1px solid #eee', borderRadius: 6, padding: 8, fontSize: 12, color: '#555' }}>
                            {(() => {
                              const info = creativeOptions.find((opt) => opt.id === v.creative_id);
                              return info?.cta_text ? (
                                <div style={{ marginBottom: 4 }}>
                                  <strong>CTA：</strong>
                                  <span>{info.cta_text}</span>
                                </div>
                              ) : (
                                <div style={{ marginBottom: 4, color: '#999' }}>无 CTA，提交后按覆盖值或创意默认</div>
                              );
                            })()}
                            {(() => {
                              const info = creativeOptions.find((opt) => opt.id === v.creative_id);
                              return info?.selling_points && info.selling_points.length > 0 ? (
                                <div>
                                  <strong>卖点：</strong>
                                  <span>{info.selling_points.slice(0, 3).join(' / ')}</span>
                                </div>
                              ) : (
                                <div style={{ color: '#999' }}>无卖点，提交后按覆盖值或创意默认</div>
                              );
                            })()}
                          </div>
                        </div>
                      )}
                      {variants.length > 2 && (
                        <button className="compact-btn compact-btn-text compact-btn-xs" onClick={() => removeVariant(idx)}>
                          删除
                        </button>
                      )}
                    </div>
                  ))}
                </div>
                <div className="compact-form-actions" style={{ marginTop: 8, display: 'flex', gap: 10 }}>
                  <button className="compact-btn compact-btn-outline compact-btn-sm" onClick={addVariant}>
                    <i className="fas fa-plus"></i>
                    <span>添加变体</span>
                  </button>
                  <button className="compact-btn compact-btn-primary" onClick={handleCreate} disabled={creating}>
                    <i className="fas fa-save"></i>
                    <span>{creating ? '创建中...' : '创建实验'}</span>
                  </button>
                  <button className="compact-btn compact-btn-text compact-btn-sm" onClick={() => navigate('/experiments')}>
                    返回列表
                  </button>
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
