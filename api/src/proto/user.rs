use crate::{proto::proto::User, router::AppError};
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

    pub async fn get_user(self, user_id: String) -> Result<User, AppError> {
        use super::proto::GetUserRequest;

        let request = GetUserRequest { user_id };

        let response = self.client.clone().get_user(request).await?;
        let inner = response.into_inner();

        Ok(inner)
    }

    pub async fn update_user(self, user_id: String, timezone: String) -> Result<User, AppError> {
        use super::proto::UpdateUserRequest;

        let request = UpdateUserRequest { user_id, timezone };

        let response = self.client.clone().update_user(request).await?;
        let inner = response.into_inner();
        Ok(inner)
    }
}
