mod proto;

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let mut auth_svc = proto::AuthSvc::new("http://[::1]:50051".to_string()).await?;

    let user = auth_svc.sign_up("Europe/London".to_string()).await?;
    println!(
        "Signed up: user_id={}, created_at={}",
        user.user_id, user.created_at
    );

    let success = auth_svc.login(user.user_id.clone()).await?;
    println!("Login success: {}", success.user_id == user.user_id);

    let mut user_svc = proto::UserSvc::new("http://[::1]:50051".to_string()).await?;
    let user = user_svc.get_user(user.user_id.clone()).await?;
    println!(
        "Got user: id={}, timezone={}, created_at={}",
        user.id, user.timezone, user.created_at
    );

    let user = user_svc
        .update_user(user.id, "America/New_York".to_string())
        .await?;
    println!(
        "Updated user: id={}, timezone={}, created_at={}",
        user.id, user.timezone, user.created_at
    );

    let mut profile_svc = proto::ProfileSvc::new("http://[::1]:50051".to_string()).await?;

    let profile = profile_svc
        .create_profile(user.id.clone(), "laptop".to_string())
        .await?;
    println!(
        "Created profile: id={}, user_id={}, name={}",
        profile.id, profile.user_id, profile.name
    );

    let profiles = profile_svc.list_profiles(user.id.clone()).await?;
    println!("Listed {} profiles", profiles.len());

    let profile = profile_svc
        .update_profile(profile.id.clone(), "desktop".to_string())
        .await?;
    println!("Updated profile: id={}, name={}", profile.id, profile.name);

    let profile = profile_svc.get_profile(profile.id.clone()).await?;
    println!("Got profile: id={}, name={}", profile.id, profile.name);

    profile_svc.delete_profile(profile.id.clone()).await?;
    println!("Deleted profile: id={}", profile.id);

    let mut whitelist_svc = proto::WhitelistSvc::new("http://[::1]:50051".to_string()).await?;
    let domains = whitelist_svc.list_permanent(profile.id.clone()).await?;
    println!("Listed {} permanent whitelists", domains.len());

    let _ = whitelist_svc
        .add_permanent(profile.id.clone(), "example.com".to_string())
        .await?;
    println!("Added permanent whitelist: example.com");

    let _ = whitelist_svc
        .remove_permanent(profile.id.clone(), "example.com".to_string())
        .await?;
    println!("Removed permanent whitelist: example.com");

    let mut category_svc = proto::CategorySvc::new("http://[::1]:50051".to_string()).await?;
    let categories = category_svc.list_blocked(profile.id.clone()).await?;
    println!("Listed {} blocked categories", categories.len());

    let _ = category_svc
        .block_category(profile.id.clone(), "porn".to_string())
        .await?;
    println!("Blocked category: porn");

    let _ = category_svc
        .unblock_category(profile.id.clone(), "porn".to_string())
        .await?;
    println!("Unblocked category: porn");

    let mut timeblock_svc = proto::TimeBlockSvc::new("http://[::1]:50051".to_string()).await?;
    let blocks = timeblock_svc.list(profile.id.clone()).await?;
    println!("Listed {} time blocks", blocks.len());

    let _ = timeblock_svc
        .create(profile.id.clone(), "social".to_string(), 0, 10, 1)
        .await?;
    println!("Created time block: social 00:00-02:30 day 1");

    let mut pool_svc = proto::PoolSvc::new("http://[::1]:50051".to_string()).await?;
    let pools = pool_svc.list_pools(user.id.clone()).await?;
    println!("Listed {} pools", pools.len());

    let pool = pool_svc
        .create_pool(
            user.id.clone(),
            "My Pool".to_string(),
            "shared".to_string(),
            1000,
        )
        .await?;
    println!(
        "Created pool: id={}, name={}, mode={}",
        pool.id, pool.name, pool.pool_mode
    );

    let _ = pool_svc
        .join_pool(pool.id.clone(), profile.id.clone())
        .await?;
    println!("Joined pool: {}", pool.id);

    let members = pool_svc.list_members(pool.id.clone()).await?;
    println!("Listed {} pool members", members.len());

    let _ = pool_svc
        .block_category(pool.id.clone(), "gambling".to_string())
        .await?;
    println!("Blocked pool category: gambling");

    let credits = pool_svc
        .get_credits(pool.id.clone(), profile.id.clone())
        .await?;
    println!(
        "Pool credits: remaining={}, total={}",
        credits.remaining, credits.total
    );

    let _ = pool_svc
        .leave_pool(pool.id.clone(), profile.id.clone())
        .await?;
    println!("Left pool: {}", pool.id);

    let _ = pool_svc
        .delete_pool(pool.id.clone(), user.id.clone())
        .await?;
    println!("Deleted pool: {}", pool.id);

    Ok(())
}
