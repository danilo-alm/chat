import { Inject, Injectable } from '@nestjs/common';
import { UserServiceClientImpl } from 'protos/ts/user/user';
import { ClientGrpc } from '@nestjs/microservices';
import { DeleteUserRequest } from 'protos/ts/gateway/gateway';
import { Empty } from 'protos/ts/google/protobuf/empty';

@Injectable()
export class DeleteUserUseCase {
  private userService: UserServiceClientImpl;

  constructor(@Inject('USER_SERVICE') private readonly client: ClientGrpc) {}

  onModuleInit() {
    this.userService =
      this.client.getService<UserServiceClientImpl>('UserService');
  }

  async execute({ id }: DeleteUserRequest): Promise<Empty> {
    await this.userService.DeleteUser({ id });
    return {};
  }
}
