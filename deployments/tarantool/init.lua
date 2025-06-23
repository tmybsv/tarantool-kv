box.cfg {
	listen = 3301
}

box.once("bootstrap", function()
	local space = box.schema.space.create("kv", { if_not_exists = true })

	space:create_index("primary", {
		type = "TREE",
		parts = { 1, "string" },
		unique = true,
		if_not_exists = true
	})

	box.schema.user.create("probeuser", {
		password = "1234qwerASDF",
		if_not_exists = true
	})
	box.schema.user.grant('probeuser', 'read,write,execute', 'universe')
end)
