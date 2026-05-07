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
 */
@Controller({ tag: 'example', prefix: '/api/example', description: 'Sample HTTP adapter wired with decorators' })
export class ExampleController {
  @Get(':id')
  @Params(exampleParamsSchema)
  @ApiResponse(200, 'Example fetched', exampleResponseSchema)
  @ApiResponse(404, 'Example not found')
  get(req: Request, res: Response): void {
    const params = req.params as unknown as ExampleParams;
    res.json({ id: params.id, name: 'sample', description: null });
  }

  @Post('/')
  @Body(exampleCreateSchema)
  @ApiResponse(201, 'Example created', exampleResponseSchema)
  create(req: Request, res: Response): void {
    const body = req.body as ExampleCreate;
    res.status(201).json({
      id: '00000000-0000-0000-0000-000000000000',
      name: body.name,
      description: body.description ?? null,
    });
  }
}
