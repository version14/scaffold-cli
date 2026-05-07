import type { ZodSchema } from 'zod';
import { getValidation, setValidation } from './metadata';

function setOn(target: object, handlerName: string, key: 'body' | 'params' | 'query', schema: ZodSchema): void {
  const meta = getValidation(target, handlerName);
  meta[key] = schema;
  setValidation(target, handlerName, meta);
}

export function Body(schema: ZodSchema): MethodDecorator {
  return (target, propertyKey) => setOn(target, propertyKey as string, 'body', schema);
}

export function Params(schema: ZodSchema): MethodDecorator {
  return (target, propertyKey) => setOn(target, propertyKey as string, 'params', schema);
}

export function Query(schema: ZodSchema): MethodDecorator {
  return (target, propertyKey) => setOn(target, propertyKey as string, 'query', schema);
}
