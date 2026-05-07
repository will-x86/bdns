fn main() -> Result<(), Box<dyn std::error::Error>> {
    tonic_prost_build::compile_protos("proto/helloworld.proto")?;
    tonic_prost_build::compile_protos("proto/auth.proto")?;
    tonic_prost_build::compile_protos("proto/user.proto")?;
    tonic_prost_build::compile_protos("proto/profile.proto")?;
    tonic_prost_build::compile_protos("proto/whitelist.proto")?;
    tonic_prost_build::compile_protos("proto/category.proto")?;
    tonic_prost_build::compile_protos("proto/timeblock.proto")?;
    tonic_prost_build::compile_protos("proto/pool.proto")?;
    Ok(())
}
