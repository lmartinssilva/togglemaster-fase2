"""Configuração de observabilidade (OpenTelemetry) do serviço.

Só é ativada quando OTEL_EXPORTER_OTLP_ENDPOINT está definida (no cluster, via
ConfigMap). Em desenvolvimento local e nos testes de CI fica desligada e não
muda o comportamento da aplicação.

Variáveis lidas automaticamente pelo SDK:
  OTEL_SERVICE_NAME            -> nome do serviço no Service Map / Grafana
  OTEL_EXPORTER_OTLP_ENDPOINT  -> http://otel-gateway.monitoring...:4318
"""
import logging
import os

log = logging.getLogger(__name__)


def init_otel(app):
    if not os.getenv("OTEL_EXPORTER_OTLP_ENDPOINT"):
        log.info("OpenTelemetry desativado (OTEL_EXPORTER_OTLP_ENDPOINT não definida)")
        return

    try:
        from opentelemetry import trace
        from opentelemetry.exporter.otlp.proto.http.trace_exporter import OTLPSpanExporter
        from opentelemetry.instrumentation.flask import FlaskInstrumentor
        from opentelemetry.instrumentation.requests import RequestsInstrumentor
        from opentelemetry.sdk.resources import Resource
        from opentelemetry.sdk.trace import TracerProvider
        from opentelemetry.sdk.trace.export import BatchSpanProcessor

        # Resource.create() já lê OTEL_SERVICE_NAME e OTEL_RESOURCE_ATTRIBUTES
        provider = TracerProvider(resource=Resource.create({}))
        provider.add_span_processor(BatchSpanProcessor(OTLPSpanExporter()))
        trace.set_tracer_provider(provider)

        # Cada request HTTP recebido vira um span (o /health é ignorado: ruído do kubelet)
        FlaskInstrumentor().instrument_app(app, excluded_urls="health")
        # Chamadas HTTP de saída (requests) viram spans e propagam o traceparent
        RequestsInstrumentor().instrument()

        try:  # só existe no analytics-service
            from opentelemetry.instrumentation.botocore import BotocoreInstrumentor
            BotocoreInstrumentor().instrument()
        except ImportError:
            pass

        log.info("OpenTelemetry ativo: service=%s", os.getenv("OTEL_SERVICE_NAME"))
    except Exception:  # observabilidade nunca pode derrubar o serviço
        log.exception("Falha ao iniciar OpenTelemetry; seguindo sem tracing")
