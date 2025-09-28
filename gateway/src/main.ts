import { NestFactory } from '@nestjs/core';
import { AppModule } from './app.module';
import {
  GrpcOptions,
  MicroserviceOptions,
  Transport,
} from '@nestjs/microservices';
import { join } from 'path';
import { ReflectionService } from '@grpc/reflection';
import { PackageDefinition } from '@grpc/proto-loader';
import { Server } from '@grpc/grpc-js';

async function bootstrap() {
  const PORT = process.env.PORT || '5000';

  const grpcOptions: GrpcOptions = {
    transport: Transport.GRPC,
    options: {
      url: `0.0.0.0:${PORT}`,
      package: 'gateway',
      protoPath: join(__dirname, '..', 'protos', 'gateway.proto'),
      onLoadPackageDefinition: (
        pkg: PackageDefinition,
        server: Pick<Server, 'addService'>,
      ) => {
        new ReflectionService(pkg).addToServer(server);
      },
    },
  };

  const app = await NestFactory.createMicroservice<MicroserviceOptions>(
    AppModule,
    grpcOptions,
  );

  await app.listen();

  console.info(`Server initialized on port ${PORT}`);
}

bootstrap().catch((error) => {
  console.error('Error initializing server:', error);
  process.exit(1);
});
