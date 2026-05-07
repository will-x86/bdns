use crate::proto::proto::User;
use tonic::transport::Channel;

use super::proto::user_service_client::UserServiceClient;

pub struct UserSvc {
    client: UserServiceClient<Channel>,
}

impl UserSvc {
    pub async fn new(addr: String) -> Result<Self, Box<dyn std::error::Error>> {
        let client = UserServiceClient::connect(addr).await?;
        Ok(Self { client })
    }

    pub async fn get_user(&mut self, user_id: String) -> Result<User, Box<dyn std::error::Error>> {
        use super::proto::GetUserRequest;

        let request = GetUserRequest { user_id };

        let response = self.client.get_user(request).await?;
        let inner = response.into_inner();

        Ok(inner)
    }

    pub async fn update_user(
        &mut self,
        user_id: String,
        timezone: String,
    ) -> Result<User, Box<dyn std::error::Error>> {
        use super::proto::UpdateUserRequest;

        let request = UpdateUserRequest { user_id, timezone };

        let response = self.client.update_user(request).await?;
        let inner = response.into_inner();
        Ok(inner)
    }
}

