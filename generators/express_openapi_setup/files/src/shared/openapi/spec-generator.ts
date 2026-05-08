import {
  OpenApiGeneratorV3,
  type OpenAPIRegistry,
  type ResponseConfig,
  type RouteConfig,
} from '@asteasolutions/zod-to-openapi';
import type { RegisteredRoute } from '../decorators';
import { getSharedRegistry } from './registry';

export interface OpenApiInfo {
  title: string;
  version: string;
  description?: string;
  [key: `x-${string}`]: unknown;
}

export interface OpenApiServer {
  url: string;
  description?: string;
  [key: `x-${string}`]: unknown;
}

export interface BuildSpecOptions {
  info: OpenApiInfo;
  servers?: OpenApiServer[];
  routes: readonly RegisteredRoute[];
  registry?: OpenAPIRegistry;
}

function pathToOpenApi(path: string): string {
  return path.replaceAll(/:([A-Za-z0-9_]+)/g, '{$1}');
}

function buildResponses(route: RegisteredRoute): RouteConfig['responses'] {
  const responses: Record<number, ResponseConfig> = {};
  for (const r of route.responses) {
    responses[r.status] = {
      description: r.description,
      ...(r.schema && {
        content: { 'application/json': { schema: r.schema } },
      }),
    };
  }
  if (Object.keys(responses).length === 0) {
    responses[200] = { description: 'Successful response' };
  }
  return responses;
}

function buildRequest(route: RegisteredRoute): RouteConfig['request'] | undefined {
  const request: NonNullable<RouteConfig['request']> = {};
  if (route.validation.params) {
    request.params = route.validation.params as NonNullable<RouteConfig['request']>['params'];
  }
  if (route.validation.query) {
    request.query = route.validation.query as NonNullable<RouteConfig['request']>['query'];
  }
  if (route.validation.body) {
    request.body = {
      content: { 'application/json': { schema: route.validation.body } },
    };
  }
  return Object.keys(request).length > 0 ? request : undefined;
}

/**
 * Convert decorator metadata into a complete OpenAPI v3 document. Pass a fresh
 * registry (`createRegistry()`) for tests, or omit to use the shared one.
 */
export function buildOpenApiSpec(options: BuildSpecOptions): Record<string, unknown> {
  const registry = options.registry ?? getSharedRegistry();

  for (const r of options.routes) {
    registry.registerPath({
      method: r.route.method,
      path: pathToOpenApi(r.fullPath),
      summary: r.route.summary,
      description: r.route.description,
      tags: [r.controller.tag],
      request: buildRequest(r),
      responses: buildResponses(r),
      security: r.protected ? [{ BearerAuth: [] }] : undefined,
    });
  }

  const generator = new OpenApiGeneratorV3(registry.definitions);
  const document = generator.generateDocument({
    openapi: '3.0.0',
    info: options.info,
    servers: options.servers,
  });
  return document as unknown as Record<string, unknown>;
}
