use crate::proto::proto::Profile;
use tonic::transport::Channel;

use super::proto::profile_service_client::ProfileServiceClient;

pub struct ProfileSvc {
    client: ProfileServiceClient<Channel>,
}

impl ProfileSvc {
    pub async fn new(addr: String) -> Result<Self, Box<dyn std::error::Error>> {
        let client = ProfileServiceClient::connect(addr).await?;
        Ok(Self { client })
    }

    pub async fn list_profiles(
        &mut self,
        user_id: String,
    ) -> Result<Vec<Profile>, Box<dyn std::error::Error>> {
        use super::proto::ListProfilesRequest;

        let request = ListProfilesRequest { user_id };

        let response = self.client.list_profiles(request).await?;
        let inner = response.into_inner();

        Ok(inner.profiles)
    }

    pub async fn create_profile(
        &mut self,
        user_id: String,
        name: String,
    ) -> Result<Profile, Box<dyn std::error::Error>> {
        use super::proto::CreateProfileRequest;

        let request = CreateProfileRequest { user_id, name };

        let response = self.client.create_profile(request).await?;
        let inner = response.into_inner();

        Ok(inner)
    }

    pub async fn get_profile(
        &mut self,
        profile_id: String,
    ) -> Result<Profile, Box<dyn std::error::Error>> {
        use super::proto::GetProfileRequest;

        let request = GetProfileRequest { profile_id };

        let response = self.client.get_profile(request).await?;
        let inner = response.into_inner();

        Ok(inner)
    }

    pub async fn update_profile(
        &mut self,
        profile_id: String,
        name: String,
    ) -> Result<Profile, Box<dyn std::error::Error>> {
        use super::proto::UpdateProfileRequest;

        let request = UpdateProfileRequest { profile_id, name };

        let response = self.client.update_profile(request).await?;
        let inner = response.into_inner();

        Ok(inner)
    }

    pub async fn delete_profile(
        &mut self,
        profile_id: String,
    ) -> Result<(), Box<dyn std::error::Error>> {
        use super::proto::DeleteProfileRequest;

        let request = DeleteProfileRequest { profile_id };

        let _response = self.client.delete_profile(request).await?;

        Ok(())
    }
}

