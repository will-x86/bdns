#[macro_export]
macro_rules! route_handlers {
    ($svc_field:ident, { $($handler_fn:ident ($req:ty => $res:ty)),* $(,)? }) => {
        $(
            pub async fn $handler_fn(
                axum::extract::State(state): axum::extract::State<$crate::router::AppState>,
                axum::Json(body): axum::Json<$req>,
            ) -> Result<axum::Json<$res>, $crate::router::AppError> {
                let resp = state.$svc_field.$handler_fn(body).await?;
                Ok(axum::Json(resp))
            }
        )*
    };
}
