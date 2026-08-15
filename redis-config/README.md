# Redis configuration

### You need to implement configuration for Redis (_redis.confg_)

    bind localhost
    port 6380
    requirepass "YourPassword"
    appendonly yes
    appendfsync everysec

 
### Also, you need to write _users.acl_ for set up users for Redis container and remove default user
    
    # Disable the default fallback user privileges
    user default off
    
    # Create an admin user with full access
    user admin on >YourPassword ~* +@all
    
    # Create a restricted app user limited to specific keys and read/write commands
    user appuser on >YourPassword ~app:* +@all -@dangerous

### When you realize steps above you should run:
    make redis