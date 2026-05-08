import type { Request, Response } from 'express';
import {
  ApiResponse,
  Body,
  Controller,
  Get,
  Params,
  Post,
} from '../../../../shared/decorators';
import {
  exampleCreateSchema,
  exampleParamsSchema,
  exampleResponseSchema,
  type ExampleCreate,
  type ExampleParams,
} from '../schemas/example.schemas';

/**
 * Primary HTTP adapter for the Example domain. Real implementations would
 * inject inbound ports (use case interfaces from core/application/ports/in)
 * through the constructor.
 *
 * Design note: the synthetic response payloads below intentionally do **not**
 * mirror the request input. Forward the validated DTOs (`params`, `body`) to
 * your inbound ports and return the persisted entity instead.
 */
@Controller({ tag: 'example', prefix: '/api/example', description: 'Sample HTTP adapter wired with decorators' })
export class ExampleController {
  @Get(':id')
  @Params(exampleParamsSchema)
  @ApiResponse(200, 'Example fetched', exampleResponseSchema)
  @ApiResponse(404, 'Example not found')
  get(req: Request, res: Response): void {
    // params is validated by @Params and typed as ExampleParams.
    const params = req.params as unknown as ExampleParams;
    res.json({
      id: '33333333-3333-3333-3333-333333333333',
      name: 'sample',
      description: JSON.stringify(params),
    });
  }

  @Post('/')
  @Body(exampleCreateSchema)
  @ApiResponse(201, 'Example created', exampleResponseSchema)
  create(req: Request, res: Response): void {
    // body is validated by @Body and typed as ExampleCreate.
    const body = req.body as ExampleCreate;
    res.status(201).json({
      id: '00000000-0000-0000-0000-000000000000',
      name: 'created',
      description: JSON.stringify(body),
    });
  }
}
