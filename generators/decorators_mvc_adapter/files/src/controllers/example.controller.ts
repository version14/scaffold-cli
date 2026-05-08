import type { Request, Response } from 'express';
import {
  ApiResponse,
  Body,
  Controller,
  Get,
  Params,
  Post,
} from '../shared/decorators';
import {
  exampleCreateSchema,
  exampleParamsSchema,
  exampleResponseSchema,
  type ExampleCreate,
  type ExampleParams,
} from '../shared/validators/example.schemas';

/**
 * Example controller — kept minimal on purpose. Replace this with real
 * controller logic and remove this file when you no longer need the
 * reference.
 *
 * Design note: the synthetic response payloads below intentionally do **not**
 * mirror the request input. Wire the validated DTOs (`params`, `body`) to
 * your model layer and return the persisted entity instead.
 */
@Controller({ tag: 'example', prefix: '/api/example', description: 'Sample resource demonstrating decorator usage' })
export class ExampleController {
  @Get(':id')
  @Params(exampleParamsSchema)
  @ApiResponse(200, 'Example fetched', exampleResponseSchema)
  @ApiResponse(404, 'Example not found')
  get(req: Request, res: Response): void {
    // params is validated by @Params and typed as ExampleParams.
    const params = req.params as unknown as ExampleParams;
    res.json({
      id: '22222222-2222-2222-2222-222222222222',
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
