box.cfg {
    listen = 3301
}
if not box.schema.user.exists('voter') then
    box.schema.user.create('voter', {password = '54321'})
end

box.schema.space.create('polls', {
    if_not_exists = true,
    format = {
        {name = 'id', type = 'string'},
        {name = 'question', type = 'string'},
        {name = 'options', type = 'array'},
        {name= 'creator_id', type = 'string'},
        {name = 'votes', type = 'array'},
        {name = 'active', type = 'string'}
    }
})

box.space.polls:create_index('primary', {
    parts = {'id'},
    if_not_exists = true
})

