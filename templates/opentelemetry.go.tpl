{{.Import}}

import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/trace"
)

{{ $decorator := (or .Vars.DecoratorName (printf "%sWithTracing" .Interface.Name)) }}

// {{$decorator}} implements {{.Interface.Type}} interface instrumented with opentracing spans
type {{$decorator}} struct {
  {{.Interface.Type}}
  _instance string
  _spanDecorator func(span trace.Span, params, results map[string]interface{})
}

// New{{$decorator}} returns {{$decorator}}
func New{{$decorator}} (base {{.Interface.Type}}, instance string, spanDecorator ...func(span trace.Span, params, results map[string]interface{})) {{$decorator}} {
  d := {{$decorator}} {
    {{.Interface.Name}}: base,
    _instance: instance,
  }

  if len(spanDecorator) > 0 && spanDecorator[0] != nil {
    d._spanDecorator = spanDecorator[0]
  }

  return d
}

{{range $method := .Interface.Methods}}
  {{if $method.AcceptsContext}}
    // {{$method.Name}} implements {{$.Interface.Type}}
func (_d {{$decorator}}) {{$method.Declaration}} {
  ctx, _span := otel.Tracer(_d._instance).Start(ctx, "{{$.Interface.Type}}.{{$method.Name}}")
  defer func() {
    if _d._spanDecorator != nil {
      _d._spanDecorator(_span, {{$method.ParamsMap}}, {{$method.ResultsMap}})
    }{{- if $method.ReturnsError}} else if err != nil {
      _span.RecordError(err)
      _span.SetAttributes(
        attribute.String("event", "error"),
        attribute.String("message", err.Error()),
      )
    }
    {{end}}
    _span.End()
  }()
  {{$method.Pass (printf "_d.%s." $.Interface.Name) }}
}
  {{end}}
{{end}}
