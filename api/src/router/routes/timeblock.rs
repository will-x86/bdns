use crate::{
    proto::proto::{
        CreateTimeBlockRequest, DeleteTimeBlockRequest, ListTimeBlocksRequest, TimeBlock,
        TimeBlockList,
    },
    route_handlers,
};

route_handlers!(timeblock_svc, {
    list (ListTimeBlocksRequest => TimeBlockList),
    create (CreateTimeBlockRequest => TimeBlock),
    delete (DeleteTimeBlockRequest => ()),
});
