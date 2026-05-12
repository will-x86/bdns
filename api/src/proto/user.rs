use crate::{
    proto::proto::{GetUserRequest, UpdateUserRequest, User},
    router::AppError,
};
use tonic::transport::Channel;

use super::proto::user_service_client::UserServiceClient;

#[derive(Clone)]
pub struct UserSvc {
    client: UserServiceClient<Channel>,
}

impl UserSvc {
    pub async fn new(addr: String) -> Result<Self, AppError> {
        let client = UserServiceClient::connect(addr).await?;
        Ok(Self { client })
    }

    pub async fn get_user(&self, req: GetUserRequest) -> Result<User, AppError> {
        let response = self.client.clone().get_user(req).await?;
        Ok(response.into_inner())
    }

    pub async fn update_user(&self, req: UpdateUserRequest) -> Result<User, AppError> {
        let response = self.client.clone().update_user(req).await?;
        Ok(response.into_inner())
    }
}
