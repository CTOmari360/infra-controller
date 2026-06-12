ALTER TABLE machines
ADD COLUMN restart_ovs_on_use_admin_network_change TEXT NOT NULL DEFAULT 'none'
CHECK (restart_ovs_on_use_admin_network_change IN ('none', 'enable', 'force_disable'));

