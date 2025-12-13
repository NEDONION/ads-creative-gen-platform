import React, { useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { creativeAPI } from '../services/api';
import Sidebar from '../components/Sidebar';
import type {
  ConfirmCopywritingRequest,
  CopywritingCandidates,
  GenerateCopywritingRequest,
  GenerateRequest,
} from '../types';

enum WorkflowStep {
  PRODUCT_INPUT = 1,
  COPYWRITING_SELECTION = 2,
  CREATIVE_CONFIG = 3,
}

const defaultFormats = ['1:1'];

const CreativeGeneratorPage: React.FC = () => {
  const navigate = useNavigate();
  const [currentStep, setCurrentStep] = useState<WorkflowStep>(WorkflowStep.PRODUCT_INPUT);

  // Step1
  const [productName, setProductName] = useState('');
  const [generatingCopywriting, setGeneratingCopywriting] = useState(false);

  // Step2
  const [candidates, setCandidates] = useState<CopywritingCandidates | null>(null);
  const [selectedCTAIndex, setSelectedCTAIndex] = useState<number>(0);
  const [selectedSPIndexes, setSelectedSPIndexes] = useState<number[]>([]);
  const [editedCTA, setEditedCTA] = useState('');
  const [editedSPs, setEditedSPs] = useState<string[]>([]);

  // Step3
  const [formData, setFormData] = useState<GenerateRequest>({
    title: '',
    selling_points: [],
    product_image_url: '',
    requested_formats: defaultFormats,
    style: '',
    cta_text: '',
    num_variants: 2,
  });
  const [submitting, setSubmitting] = useState(false);

  const canProceedToCopywriting = useMemo(() => productName.trim().length > 0, [productName]);

  const handleGenerateCopywriting = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!canProceedToCopywriting) return;
    setGeneratingCopywriting(true);
    try {
      const req: GenerateCopywritingRequest = { product_name: productName.trim() };
      const res = await creativeAPI.generateCopywriting(req);
      if (res.code === 0 && res.data) {
        setCandidates(res.data);
        setSelectedCTAIndex(0);
        setSelectedSPIndexes([0]);
        setEditedCTA('');
        setEditedSPs([]);
        setFormData((prev) => ({ ...prev, title: productName.trim() }));
        setCurrentStep(WorkflowStep.COPYWRITING_SELECTION);
      } else {
        alert(res.message || '生成文案失败');
      }
    } catch (err) {
      console.error(err);
      alert('生成文案失败: ' + (err as Error).message);
    } finally {
      setGeneratingCopywriting(false);
    }
  };

  const handleConfirmCopywriting = async () => {
    if (!candidates) return;
    if (selectedSPIndexes.length === 0 && editedSPs.length === 0) {
      alert('请至少选择一个卖点');
      return;
    }
    setSubmitting(true);
    try {
      const payload: ConfirmCopywritingRequest = {
        task_id: candidates.task_id,
        selected_cta_index: selectedCTAIndex,
        selected_sp_indexes: selectedSPIndexes,
        edited_cta: editedCTA || undefined,
        edited_sps: editedSPs.length > 0 ? editedSPs : undefined,
        product_image_url: formData.product_image_url || undefined,
        style: formData.style || undefined,
        num_variants: formData.num_variants,
        formats: formData.requested_formats,
      };
      const res = await creativeAPI.confirmCopywriting(payload);
      if (res.code !== 0) {
        alert(res.message || '确认文案失败');
        return;
      }
      setCurrentStep(WorkflowStep.CREATIVE_CONFIG);
    } catch (err) {
      console.error(err);
      alert('确认文案失败: ' + (err as Error).message);
    } finally {
      setSubmitting(false);
    }
  };

  const handleStartCreative = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!candidates) {
      alert('请先生成并确认文案');
      return;
    }
    setSubmitting(true);
    try {
      const res = await creativeAPI.startCreative({ task_id: candidates.task_id });
      if (res.code === 0 && res.data) {
        alert(`创意生成已开始！任务ID: ${res.data.task_id}`);
        navigate('/tasks');
      } else {
        alert(res.message || '启动创意生成失败');
      }
    } catch (err) {
      console.error(err);
      alert('启动创意生成失败: ' + (err as Error).message);
    } finally {
      setSubmitting(false);
    }
  };

  const toggleSellingPoint = (idx: number) => {
    setSelectedSPIndexes((prev) =>
      prev.includes(idx) ? prev.filter((i) => i !== idx) : [...prev, idx]
    );
  };

  const resetAll = () => {
    setProductName('');
    setCandidates(null);
    setSelectedCTAIndex(0);
    setSelectedSPIndexes([]);
    setEditedCTA('');
    setEditedSPs([]);
    setFormData({
      title: '',
      selling_points: [],
      product_image_url: '',
      requested_formats: defaultFormats,
      style: '',
      cta_text: '',
      num_variants: 2,
    });
    setCurrentStep(WorkflowStep.PRODUCT_INPUT);
  };

  const renderStepIndicator = () => (
    <div className="step-indicator">
      {[WorkflowStep.PRODUCT_INPUT, WorkflowStep.COPYWRITING_SELECTION, WorkflowStep.CREATIVE_CONFIG].map((step, idx) => {
        const titles = ['商品', '文案', '生成'];
        return (
          <div key={step} className={`step ${currentStep === step ? 'active' : ''} ${currentStep > step ? 'done' : ''}`}>
            <div className="step-number">{idx + 1}</div>
            <div className="step-label">{titles[idx]}</div>
          </div>
        );
      })}
    </div>
  );

  return (
    <div className="app">
      <Sidebar />

      <div className="main-content">
        <div className="header">
          <h1 className="page-title">文案 + 创意生成</h1>
          <div className="user-info">
            <div className="avatar">A</div>
            <span>管理员</span>
          </div>
        </div>

        <div className="content">
          <div className="compact-layout">
            {renderStepIndicator()}

            {/* Step 1: 商品输入 */}
            {currentStep === WorkflowStep.PRODUCT_INPUT && (
              <div className="compact-card">
                <div className="compact-card-header">
                  <h3 className="compact-card-title">步骤 1：输入商品</h3>
                  <div className="compact-card-hint">输入商品名称，系统自动生成 CTA 和卖点</div>
                </div>
                <div className="compact-card-body">
                  <form onSubmit={handleGenerateCopywriting}>
                    <div className="compact-form-group full-width">
                      <label className="compact-label">
                        <span className="label-text">商品名称</span>
                        <span className="label-required">*</span>
                      </label>
                      <input
                        type="text"
                        className="compact-input"
                        value={productName}
                        onChange={(e) => setProductName(e.target.value)}
                        placeholder="如：智能手表 Pro"
                        required
                      />
                    </div>
                    <div className="compact-form-actions">
                      <button type="submit" className="compact-btn compact-btn-primary" disabled={!canProceedToCopywriting || generatingCopywriting}>
                        <i className="fas fa-magic"></i>
                        <span>{generatingCopywriting ? '生成中...' : '生成文案'}</span>
                      </button>
                    </div>
                  </form>
                </div>
              </div>
            )}

            {/* Step 2: 文案选择 */}
            {currentStep === WorkflowStep.COPYWRITING_SELECTION && candidates && (
              <div className="compact-card">
                <div className="compact-card-header">
                  <h3 className="compact-card-title">步骤 2：选择/编辑文案</h3>
                  <div className="compact-card-hint">选择喜欢的 CTA 和卖点，可手动微调</div>
                </div>
                <div className="compact-card-body">
                  <div className="compact-section-title">CTA（行动号召）</div>
                  <div className="option-grid">
                    {candidates.cta_candidates.map((cta, idx) => (
                      <label key={idx} className={`radio-option ${selectedCTAIndex === idx ? 'active' : ''}`}>
                        <input
                          type="radio"
                          name="cta"
                          checked={selectedCTAIndex === idx}
                          onChange={() => setSelectedCTAIndex(idx)}
                        />
                        <span className="radio-label">{cta}</span>
                      </label>
                    ))}
                  </div>
                  <div className="compact-form-group full-width">
                    <label className="compact-label">
                      <span className="label-text">编辑 CTA（可选）</span>
                    </label>
                    <input
                      type="text"
                      className="compact-input"
                      placeholder="不填则使用选择的 CTA"
                      value={editedCTA}
                      onChange={(e) => setEditedCTA(e.target.value)}
                    />
                  </div>

                  <div className="compact-section-title">卖点（至少选一项）</div>
                  <div className="option-grid">
                    {candidates.selling_point_candidates.map((sp, idx) => (
                      <label key={idx} className={`checkbox-option ${selectedSPIndexes.includes(idx) ? 'active' : ''}`}>
                        <input
                          type="checkbox"
                          checked={selectedSPIndexes.includes(idx)}
                          onChange={() => toggleSellingPoint(idx)}
                        />
                        <span className="checkbox-label">{sp}</span>
                      </label>
                    ))}
                  </div>
                  <div className="compact-form-group full-width">
                    <label className="compact-label">
                      <span className="label-text">编辑卖点（可选，一行一个）</span>
                    </label>
                    <textarea
                      className="compact-textarea"
                      rows={3}
                      placeholder="不填则使用勾选的卖点"
                      value={editedSPs.join('\n')}
                      onChange={(e) => setEditedSPs(e.target.value.split('\n').map((v) => v.trim()).filter(Boolean))}
                    ></textarea>
                  </div>

                  <div className="compact-form-actions">
                    <button type="button" className="compact-btn compact-btn-outline" onClick={() => setCurrentStep(WorkflowStep.PRODUCT_INPUT)}>
                      <i className="fas fa-arrow-left"></i>
                      <span>返回</span>
                    </button>
                    <button type="button" className="compact-btn compact-btn-primary" onClick={handleConfirmCopywriting} disabled={submitting}>
                      <i className="fas fa-check"></i>
                      <span>{submitting ? '提交中...' : '确认文案'}</span>
                    </button>
                  </div>
                </div>
              </div>
            )}

            {/* Step 3: 创意配置与启动 */}
            {currentStep === WorkflowStep.CREATIVE_CONFIG && candidates && (
              <div className="compact-card">
                <div className="compact-card-header">
                  <h3 className="compact-card-title">步骤 3：生成创意</h3>
                  <div className="compact-card-hint">设置风格/图片，提交即可生成</div>
                </div>
                <div className="compact-card-body">
                  <form onSubmit={handleStartCreative}>
                    <div className="compact-form-grid">
                      <div className="compact-form-group">
                        <label className="compact-label">
                          <span className="label-text">创意风格</span>
                        </label>
                        <select
                          className="compact-input"
                          value={formData.style}
                          onChange={(e) => setFormData({ ...formData, style: e.target.value })}
                        >
                          <option value="">通用风格</option>
                          <option value="bright">明亮风格</option>
                          <option value="professional">专业风格</option>
                          <option value="modern">现代风格</option>
                          <option value="elegant">优雅风格</option>
                        </select>
                      </div>

                      <div className="compact-form-group">
                        <label className="compact-label">
                          <span className="label-text">变体数量</span>
                        </label>
                        <input
                          type="number"
                          className="compact-input"
                          value={formData.num_variants}
                          onChange={(e) => setFormData({ ...formData, num_variants: parseInt(e.target.value) || 3 })}
                          min={1}
                          max={10}
                        />
                      </div>

                      <div className="compact-form-group">
                        <label className="compact-label">
                          <span className="label-text">产品图片URL（可选）</span>
                        </label>
                        <input
                          type="url"
                          className="compact-input"
                          value={formData.product_image_url}
                          onChange={(e) => setFormData({ ...formData, product_image_url: e.target.value })}
                          placeholder="https://example.com/product.jpg"
                        />
                      </div>

                      <div className="compact-form-group">
                        <label className="compact-label">
                          <span className="label-text">尺寸</span>
                        </label>
                        <input
                          type="text"
                          className="compact-input"
                          value={formData.requested_formats.join(',')}
                          onChange={(e) =>
                            setFormData({
                              ...formData,
                              requested_formats: e.target.value
                                .split(',')
                                .map((v) => v.trim())
                                .filter(Boolean),
                            })
                          }
                          placeholder="例如: 1:1,9:16"
                        />
                      </div>
                    </div>

                    <div className="compact-form-actions">
                      <button type="button" className="compact-btn compact-btn-outline" onClick={() => setCurrentStep(WorkflowStep.COPYWRITING_SELECTION)}>
                        <i className="fas fa-arrow-left"></i>
                        <span>返回</span>
                      </button>
                      <button type="submit" className="compact-btn compact-btn-primary" disabled={submitting}>
                        <i className="fas fa-bolt"></i>
                        <span>{submitting ? '提交中...' : '开始生成'}</span>
                      </button>
                    </div>
                  </form>
                </div>
              </div>
            )}

            <div className="compact-form-actions" style={{ marginTop: '8px' }}>
              <button type="button" className="compact-btn compact-btn-text" onClick={resetAll}>
                <i className="fas fa-undo"></i>
                <span>重新开始</span>
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default CreativeGeneratorPage;
