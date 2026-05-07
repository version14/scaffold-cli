import { setController, type ControllerMetadata } from './metadata';

export interface ControllerOptions {
  tag: string;
  prefix?: string;
  description?: string;
}

/**
 * Marks a class as an HTTP controller. The `tag` is used as the OpenAPI group;
 * the `prefix` is prepended to every route path declared on the class.
 */
export function Controller(options: ControllerOptions): ClassDecorator {
  return (target) => {
    const meta: ControllerMetadata = {
      tag: options.tag,
      prefix: options.prefix ?? '',
      description: options.description,
    };
    setController(target.prototype, meta);
  };
}
