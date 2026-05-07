import 'reflect-metadata';
import type { ZodSchema } from 'zod';

export const META_KEYS = {
  CONTROLLER: Symbol('decorators:controller'),
  ROUTES: Symbol('decorators:routes'),
  VALIDATION: Symbol('decorators:validation'),
  RESPONSES: Symbol('decorators:responses'),
  AUTH: Symbol('decorators:auth'),
  REQUIRED_HEADERS: Symbol('decorators:required-headers'),
} as const;

export type HttpMethod = 'get' | 'post' | 'put' | 'patch' | 'delete';

export interface ControllerMetadata {
  tag: string;
  prefix: string;
  description?: string;
}

export interface RouteMetadata {
  method: HttpMethod;
  path: string;
  handlerName: string;
  summary?: string;
  description?: string;
}

export interface ValidationMetadata {
  body?: ZodSchema;
  params?: ZodSchema;
  query?: ZodSchema;
}

export interface ResponseMetadata {
  status: number;
  description: string;
  schema?: ZodSchema;
}

export interface RequiredHeadersMetadata {
  headers: string[];
}

export function getRoutes(target: object): RouteMetadata[] {
  return (Reflect.getMetadata(META_KEYS.ROUTES, target) as RouteMetadata[]) ?? [];
}

export function addRoute(target: object, route: RouteMetadata): void {
  const routes = getRoutes(target);
  routes.push(route);
  Reflect.defineMetadata(META_KEYS.ROUTES, routes, target);
}

export function getController(target: object): ControllerMetadata | undefined {
  return Reflect.getMetadata(META_KEYS.CONTROLLER, target) as ControllerMetadata | undefined;
}

export function setController(target: object, meta: ControllerMetadata): void {
  Reflect.defineMetadata(META_KEYS.CONTROLLER, meta, target);
}

export function getValidation(target: object, handlerName: string | symbol): ValidationMetadata {
  return (Reflect.getMetadata(META_KEYS.VALIDATION, target, handlerName as string) as ValidationMetadata) ?? {};
}

export function setValidation(target: object, handlerName: string | symbol, meta: ValidationMetadata): void {
  Reflect.defineMetadata(META_KEYS.VALIDATION, meta, target, handlerName as string);
}

export function getResponses(target: object, handlerName: string | symbol): ResponseMetadata[] {
  return (Reflect.getMetadata(META_KEYS.RESPONSES, target, handlerName as string) as ResponseMetadata[]) ?? [];
}

export function addResponse(target: object, handlerName: string | symbol, response: ResponseMetadata): void {
  const existing = getResponses(target, handlerName);
  existing.push(response);
  Reflect.defineMetadata(META_KEYS.RESPONSES, existing, target, handlerName as string);
}

export function isProtected(target: object, handlerName: string | symbol): boolean {
  return Reflect.getMetadata(META_KEYS.AUTH, target, handlerName as string) === true;
}

export function setProtected(target: object, handlerName: string | symbol): void {
  Reflect.defineMetadata(META_KEYS.AUTH, true, target, handlerName as string);
}

export function getRequiredHeaders(target: object, handlerName: string | symbol): string[] {
  const meta = Reflect.getMetadata(META_KEYS.REQUIRED_HEADERS, target, handlerName as string) as
    | RequiredHeadersMetadata
    | undefined;
  return meta?.headers ?? [];
}

export function setRequiredHeaders(target: object, handlerName: string | symbol, headers: string[]): void {
  Reflect.defineMetadata(
    META_KEYS.REQUIRED_HEADERS,
    { headers: headers.map((h) => h.toLowerCase()) },
    target,
    handlerName as string,
  );
}
