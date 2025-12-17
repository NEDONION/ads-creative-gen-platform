package cache

import "encoding/json"

// JSONCodec 使用 encoding/json，保证与现有结构的兼容性
type JSONCodec struct{}

func (JSONCodec) Marshal(v any) ([]byte, error)      { return json.Marshal(v) }
func (JSONCodec) Unmarshal(data []byte, v any) error { return json.Unmarshal(data, v) }
