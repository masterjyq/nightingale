// Code generated by easyjson for marshaling/unmarshaling. DO NOT EDIT.

package router

import (
	json "encoding/json"
	easyjson "github.com/mailru/easyjson"
	jlexer "github.com/mailru/easyjson/jlexer"
	jwriter "github.com/mailru/easyjson/jwriter"
)

// suppress unused package warning
var (
	_ *json.RawMessage
	_ *jlexer.Lexer
	_ *jwriter.Writer
	_ easyjson.Marshaler
)

func easyjson61ba9b47DecodeGithubComDidiNightingaleV5SrcServerRouter(in *jlexer.Lexer, out *FalconMetricArr) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		in.Skip()
		*out = nil
	} else {
		in.Delim('[')
		if *out == nil {
			if !in.IsDelim(']') {
				*out = make(FalconMetricArr, 0, 0)
			} else {
				*out = FalconMetricArr{}
			}
		} else {
			*out = (*out)[:0]
		}
		for !in.IsDelim(']') {
			var v1 FalconMetric
			(v1).UnmarshalEasyJSON(in)
			*out = append(*out, v1)
			in.WantComma()
		}
		in.Delim(']')
	}
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson61ba9b47EncodeGithubComDidiNightingaleV5SrcServerRouter(out *jwriter.Writer, in FalconMetricArr) {
	if in == nil && (out.Flags&jwriter.NilSliceAsEmpty) == 0 {
		out.RawString("null")
	} else {
		out.RawByte('[')
		for v2, v3 := range in {
			if v2 > 0 {
				out.RawByte(',')
			}
			(v3).MarshalEasyJSON(out)
		}
		out.RawByte(']')
	}
}

// MarshalJSON supports json.Marshaler interface
func (v FalconMetricArr) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson61ba9b47EncodeGithubComDidiNightingaleV5SrcServerRouter(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v FalconMetricArr) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson61ba9b47EncodeGithubComDidiNightingaleV5SrcServerRouter(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *FalconMetricArr) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson61ba9b47DecodeGithubComDidiNightingaleV5SrcServerRouter(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *FalconMetricArr) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson61ba9b47DecodeGithubComDidiNightingaleV5SrcServerRouter(l, v)
}
func easyjson61ba9b47DecodeGithubComDidiNightingaleV5SrcServerRouter1(in *jlexer.Lexer, out *FalconMetric) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "metric":
			out.Metric = string(in.String())
		case "endpoint":
			out.Endpoint = string(in.String())
		case "timestamp":
			out.Timestamp = int64(in.Int64())
		case "value":
			if m, ok := out.ValueUnTyped.(easyjson.Unmarshaler); ok {
				m.UnmarshalEasyJSON(in)
			} else if m, ok := out.ValueUnTyped.(json.Unmarshaler); ok {
				_ = m.UnmarshalJSON(in.Raw())
			} else {
				out.ValueUnTyped = in.Interface()
			}
		case "tags":
			out.Tags = string(in.String())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson61ba9b47EncodeGithubComDidiNightingaleV5SrcServerRouter1(out *jwriter.Writer, in FalconMetric) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"metric\":"
		out.RawString(prefix[1:])
		out.String(string(in.Metric))
	}
	{
		const prefix string = ",\"endpoint\":"
		out.RawString(prefix)
		out.String(string(in.Endpoint))
	}
	{
		const prefix string = ",\"timestamp\":"
		out.RawString(prefix)
		out.Int64(int64(in.Timestamp))
	}
	{
		const prefix string = ",\"value\":"
		out.RawString(prefix)
		if m, ok := in.ValueUnTyped.(easyjson.Marshaler); ok {
			m.MarshalEasyJSON(out)
		} else if m, ok := in.ValueUnTyped.(json.Marshaler); ok {
			out.Raw(m.MarshalJSON())
		} else {
			out.Raw(json.Marshal(in.ValueUnTyped))
		}
	}
	{
		const prefix string = ",\"tags\":"
		out.RawString(prefix)
		out.String(string(in.Tags))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v FalconMetric) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson61ba9b47EncodeGithubComDidiNightingaleV5SrcServerRouter1(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v FalconMetric) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson61ba9b47EncodeGithubComDidiNightingaleV5SrcServerRouter1(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *FalconMetric) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson61ba9b47DecodeGithubComDidiNightingaleV5SrcServerRouter1(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *FalconMetric) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson61ba9b47DecodeGithubComDidiNightingaleV5SrcServerRouter1(l, v)
}