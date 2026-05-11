use crate::{proto::proto::Profile, router::AppError};
use tonic::transport::Channel;

use super::proto::profile_service_client::ProfileServiceClient;

#[derive(Clone)]
pub struct ProfileSvc {
    client: ProfileServiceClient<Channel>,
}

impl ProfileSvc {
    pub async fn new(addr: String) -> Result<Self, AppError> {
        let client = ProfileServiceClient::connect(addr).await?;
        Ok(Self { client })
    }

    pub async fn list_profiles(self, user_id: String) -> Result<Vec<Profile>, AppError> {
        use super::proto::ListProfilesRequest;

        let request = ListProfilesRequest { user_id };

        let response = self.client.clone().list_profiles(request).await?;
        let inner = response.into_inner();

        Ok(inner.profiles)
    }

    pub async fn create_profile(self, user_id: String, name: String) -> Result<Profile, AppError> {
        use super::proto::CreateProfileRequest;

        let request = CreateProfileRequest { user_id, name };

        let response = self.client.clone().create_profile(request).await?;
        let inner = response.into_inner();

        Ok(inner)
    }

    pub async fn get_profile(self, profile_id: String) -> Result<Profile, AppError> {
        use super::proto::GetProfileRequest;

        let request = GetProfileRequest { profile_id };

        let response = self.client.clone().get_profile(request).await?;
        let inner = response.into_inner();

        Ok(inner)
    }

    pub async fn update_profile(
        self,
        profile_id: String,
        name: String,
    ) -> Result<Profile, AppError> {
        use super::proto::UpdateProfileRequest;

        let request = UpdateProfileRequest { profile_id, name };

        let response = self.client.clone().update_profile(request).await?;
        let inner = response.into_inner();

        Ok(inner)
    }

    pub async fn delete_profile(self, profile_id: String) -> Result<(), AppError> {
        use super::proto::DeleteProfileRequest;

        let request = DeleteProfileRequest { profile_id };

        let _response = self.client.clone().delete_profile(request).await?;

        Ok(())
    }
}
