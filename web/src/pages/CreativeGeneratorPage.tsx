import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import Layout from '../components/Layout';
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
    <Layout title="创意生成">
      <div className="card">
        <div className="card-header">
          <h3 className="card-title">生成创意</h3>
        </div>
        <div className="card-body">
          <form onSubmit={handleSubmit}>
            <div className="form-row">
              <div className="form-group">
                <label className="form-label">创意标题 *</label>
                <input
                  type="text"
                  className="form-control"
                  value={formData.title}
                  onChange={(e) => setFormData({ ...formData, title: e.target.value })}
                  required
                  placeholder="输入创意标题"
                />
              </div>
              <div className="form-group">
                <label className="form-label">创意风格</label>
                <select
                  className="form-control"
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
            </div>

            <div className="form-group">
              <label className="form-label">核心卖点 *</label>
              <textarea
                className="form-control"
                value={sellingPointsText}
                onChange={(e) => setSellingPointsText(e.target.value)}
                required
                placeholder="每行一个卖点，例如：5折优惠、限时抢购"
              ></textarea>
            </div>

            <div className="form-row">
              <div className="form-group">
                <label className="form-label">变体数量</label>
                <input
                  type="number"
                  className="form-control"
                  value={formData.num_variants}
                  onChange={(e) => setFormData({ ...formData, num_variants: parseInt(e.target.value) || 3 })}
                  min="1"
                  max="10"
                />
              </div>
            </div>

            <div className="form-group">
              <label className="form-label">产品图片URL</label>
              <input
                type="url"
                className="form-control"
                value={formData.product_image_url}
                onChange={(e) => setFormData({ ...formData, product_image_url: e.target.value })}
                placeholder="https://example.com/product.jpg"
              />
            </div>

            <div className="form-group">
              <label className="form-label">行动号召文字</label>
              <input
                type="text"
                className="form-control"
                value={formData.cta_text}
                onChange={(e) => setFormData({ ...formData, cta_text: e.target.value })}
                placeholder="立即购买、了解更多等"
              />
            </div>

            <button type="submit" className="btn btn-primary" disabled={submitting}>
              <i className="fas fa-bolt"></i> {submitting ? '生成中...' : '生成创意'}
            </button>
            <button type="button" className="btn btn-outline" onClick={handleReset} style={{ marginLeft: '12px' }}>
              <i className="fas fa-redo"></i> 重置
            </button>
          </form>
        </div>
      </div>
    </Layout>
  );
};

export default CreativeGeneratorPage;
