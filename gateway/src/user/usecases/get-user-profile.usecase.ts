import {
  BadRequestException,
  Inject,
  Injectable,
  InternalServerErrorException,
} from '@nestjs/common';
import { GetUserResponse, UserServiceClientImpl } from 'protos/ts/user/user';
import { ClientGrpc, RpcException } from '@nestjs/microservices';
import {
  GetUserProfileRequest,
  GetUserProfileResponse,
} from 'protos/ts/gateway/gateway';

@Injectable()
export class GetUserProfileUseCase {
  private userService: UserServiceClientImpl;

  constructor(@Inject('USER_SERVICE') private readonly client: ClientGrpc) {}

  onModuleInit() {
    this.userService =
      this.client.getService<UserServiceClientImpl>('UserService');
  }

  async execute({
    id,
    username,
  }: GetUserProfileRequest): Promise<GetUserProfileResponse> {
    console.log('GetUserProfileUseCase.execute called with:', { id, username });
    let userProfilePromise: Promise<GetUserResponse>;

    if (id) {
      userProfilePromise = this.userService.GetUserById({ id });
    } else if (username) {
      userProfilePromise = this.userService.GetUserByUsername({ username });
    } else {
      throw new RpcException(
        new BadRequestException('Id or username is required'),
      );
    }

    const userProfile = (await userProfilePromise).user;
    console.log('User profile retrieved:', userProfile);
    if (!userProfile) {
      throw new RpcException(
        new InternalServerErrorException('Could not retrieve user profile'),
      );
    }

    return {
      id: userProfile.id,
      name: userProfile.name,
      username: userProfile.username,
    };
  }
}
