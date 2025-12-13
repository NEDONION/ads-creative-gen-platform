import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { creativeAPI } from '../services/api';
import type { GenerateRequest } from '../types';

const CreativeGeneratorPage: React.FC = () => {
  const navigate = useNavigate();
  const [formData, setFormData] = useState<GenerateRequest>({
    title: '',
    selling_points: [],
    product_image_url: '',
    requested_formats: ['1:1'],
    style: '',
    cta_text: '',
    num_variants: 3,
  });
  const [sellingPointsText, setSellingPointsText] = useState('');
  const [submitting, setSubmitting] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitting(true);

    try {
      const sellingPoints = sellingPointsText
        .split('\n')
        .map((point) => point.trim())
        .filter((point) => point !== '');

      const requestData: GenerateRequest = {
        ...formData,
        selling_points: sellingPoints,
      };

      const response = await creativeAPI.generate(requestData);

      if (response.code === 0 && response.data) {
        alert(`创意生成请求已提交！任务ID: ${response.data.task_id}`);
        navigate('/tasks');
      } else {
        alert('生成失败: ' + response.message);
      }
    } catch (err) {
      alert('请求失败: ' + (err as Error).message);
      console.error('Generate error:', err);
    } finally {
      setSubmitting(false);
    }
  };

  const handleReset = () => {
    setFormData({
      title: '',
      selling_points: [],
      product_image_url: '',
      requested_formats: ['1:1'],
      style: '',
      cta_text: '',
      num_variants: 3,
    });
    setSellingPointsText('');
  };

  return (
    <div className="app">
      <div className="sidebar">
        <div className="sidebar-header">
          <h2>
            <i className="fas fa-bullseye"></i> <span>创意平台</span>
          </h2>
        </div>
        <nav className="nav-menu">
          <a href="/" className="nav-item">
            <i className="fas fa-home"></i>
            <span>仪表盘</span>
          </a>
          <a href="/creative" className="nav-item active">
            <i className="fas fa-magic"></i>
            <span>创意生成</span>
          </a>
          <a href="/assets" className="nav-item">
            <i className="fas fa-images"></i>
            <span>素材管理</span>
          </a>
          <a href="/tasks" className="nav-item">
            <i className="fas fa-tasks"></i>
            <span>任务管理</span>
          </a>
        </nav>
      </div>

      <div className="main-content">
        <div className="header">
          <h1 className="page-title">创意生成</h1>
          <div className="user-info">
            <div className="avatar">A</div>
            <span>管理员</span>
          </div>
        </div>

        <div className="content">
          <div className="compact-layout">
            <div className="compact-card">
              <div className="compact-card-header">
                <h3 className="compact-card-title">生成新创意</h3>
                <div className="compact-card-hint">填写以下信息生成AI驱动的广告创意</div>
              </div>
              <div className="compact-card-body">
                <form onSubmit={handleSubmit}>
                  <div className="compact-form-grid">
                    <div className="compact-form-group">
                      <label className="compact-label">
                        <span className="label-text">创意标题</span>
                        <span className="label-required">*</span>
                      </label>
                      <input
                        type="text"
                        className="compact-input"
                        value={formData.title}
                        onChange={(e) => setFormData({ ...formData, title: e.target.value })}
                        required
                        placeholder="输入创意标题"
                      />
                    </div>

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
                        min="1"
                        max="10"
                      />
                    </div>

                    <div className="compact-form-group">
                      <label className="compact-label">
                        <span className="label-text">产品图片URL</span>
                      </label>
                      <input
                        type="url"
                        className="compact-input"
                        value={formData.product_image_url}
                        onChange={(e) => setFormData({ ...formData, product_image_url: e.target.value })}
                        placeholder="https://example.com/product.jpg"
                      />
                    </div>

                    <div className="compact-form-group full-width">
                      <label className="compact-label">
                        <span className="label-text">核心卖点</span>
                        <span className="label-required">*</span>
                      </label>
                      <textarea
                        className="compact-textarea"
                        value={sellingPointsText}
                        onChange={(e) => setSellingPointsText(e.target.value)}
                        required
                        placeholder="每行一个卖点，例如：&#10;• 5折优惠&#10;• 限时抢购&#10;• 包邮到家"
                        rows={4}
                      ></textarea>
                    </div>

                    <div className="compact-form-group full-width">
                      <label className="compact-label">
                        <span className="label-text">行动号召文字</span>
                      </label>
                      <input
                        type="text"
                        className="compact-input"
                        value={formData.cta_text}
                        onChange={(e) => setFormData({ ...formData, cta_text: e.target.value })}
                        placeholder="立即购买、了解更多等"
                      />
                    </div>
                  </div>

                  <div className="compact-form-actions">
                    <button type="submit" className="compact-btn compact-btn-primary" disabled={submitting}>
                      <i className="fas fa-bolt"></i>
                      <span>{submitting ? '生成中...' : '生成创意'}</span>
                    </button>
                    <button type="button" className="compact-btn compact-btn-outline" onClick={handleReset}>
                      <i className="fas fa-redo"></i>
                      <span>重置</span>
                    </button>
                  </div>
                </form>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default CreativeGeneratorPage;
