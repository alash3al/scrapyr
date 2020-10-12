scrapyr
========
> a very simple scrapy orchestrator engine that could be distributed among multiple machines to build a scrapy cluster, under-the-hood it uses redis as a task broker, it may be changed in the future to support pluggable brokers, but for now it does the job.

Features
========
- uses simple configuration language for humans called `hcl`.
- multiple types of queues/workers (`lifo`, `fifo`, `weight`).
- you can define multiple workers with different type of queues.
- abbility to override the content of the `settings.py` of the scrapy project from the same configuration file.
- a `status` endpoint helps you to understand what is going on.
- a `enqueue` endpoint lets you push a job into the specified queue, as well the abbility to execute the job instantly and returns the extracted items.

API Examples
============

- Getting the status of the cluster
```bash
curl --request GET \
  --url http://localhost:1993/status \
  --header 'content-type: application/json'
```

- Push a task into the queue utilizing the worker `worker1` which is pre-defined in the `scrapyr.hcl`
```bash
# worker -> the worker name (predefined in scrapyr.hcl)
# spider -> the scrapy spider to be executed
# max_execution_time -> the max duration the scrapy process should take
# args -> a key value strings will be translated to `-a key=value ...` for each key-value pair.
# weight -> the weight of the task itself (in case of weight based workers defined in the scrapyr.hcl)
curl --request POST \
  --url http://localhost:1993/enqueue \
  --header 'content-type: application/json' \
  --data '{
	"worker": "worker1",
	"spider": "spider_name",
	"max_execution_time": "20s",
	"args": {
            "scrapy_arg_name": "scrapy_arg_value"
      },
	"weight": 10
}'
```

Configurations
===============
> here is an example of the `scraply.hcl`

```hcl

# the webserver listening address
listen_addr = ":1993"

# redis connection string
# it uses url-style connection string
# example: redis://username:password@hostname:port/database_number
redis_dsn = "redis://127.0.0.1:6378/1"

scrapy {
    project_dir = "${HOME}/playground/tstscrapy"

    python_bin = "/usr/bin/python3"

    items_dir = "${PWD}/data"
}

worker worker1 {
    // which method you want the worker to use
    // lifo: last in, first out
    // fifo: first in, first out
    // weight: max weight, first out
    use = "weight"

    // max processes to be executed in the same time for this workers
    max_procs = 5
}


# sometimes you may need to control the `ProjectNAme/ProjectName/settings.py` file from here
# so we did this special key which mounts the contents of it into `settings.py` file.
settings_py = <<PYTHON
# Scrapy settings for tstscrapy project
#
# For simplicity, this file contains only settings considered important or
# commonly used. You can find more settings consulting the documentation:
#
#     https://docs.scrapy.org/en/latest/topics/settings.html
#     https://docs.scrapy.org/en/latest/topics/downloader-middleware.html
#     https://docs.scrapy.org/en/latest/topics/spider-middleware.html

BOT_NAME = 'tstscrapy'

SPIDER_MODULES = ['tstscrapy.spiders']
NEWSPIDER_MODULE = 'tstscrapy.spiders'


# Crawl responsibly by identifying yourself (and your website) on the user-agent
#USER_AGENT = 'tstscrapy (+http://www.yourdomain.com)'

# Obey robots.txt rules
ROBOTSTXT_OBEY = False

# Configure maximum concurrent requests performed by Scrapy (default: 16)
#CONCURRENT_REQUESTS = 32

# Configure a delay for requests for the same website (default: 0)
# See https://docs.scrapy.org/en/latest/topics/settings.html#download-delay
# See also autothrottle settings and docs
#DOWNLOAD_DELAY = 3
# The download delay setting will honor only one of:
#CONCURRENT_REQUESTS_PER_DOMAIN = 16
#CONCURRENT_REQUESTS_PER_IP = 16

# Disable cookies (enabled by default)
#COOKIES_ENABLED = False

# Disable Telnet Console (enabled by default)
#TELNETCONSOLE_ENABLED = False

# Override the default request headers:
# DEFAULT_REQUEST_HEADERS = {
#   'Accept': 'text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8',
#   'Accept-Language': 'en',
# }

# Enable or disable spider middlewares
# See https://docs.scrapy.org/en/latest/topics/spider-middleware.html
# SPIDER_MIDDLEWARES = {
#    'tstscrapy.middlewares.TstscrapySpiderMiddleware': 543,
# }

# Enable or disable downloader middlewares
# See https://docs.scrapy.org/en/latest/topics/downloader-middleware.html
# DOWNLOADER_MIDDLEWARES = {
#    'tstscrapy.middlewares.TstscrapyDownloaderMiddleware': 543,
# }

# Enable or disable extensions
# See https://docs.scrapy.org/en/latest/topics/extensions.html
# EXTENSIONS = {
#    'scrapy.extensions.telnet.TelnetConsole': None,
# }

# Configure item pipelines
# See https://docs.scrapy.org/en/latest/topics/item-pipeline.html
# ITEM_PIPELINES = {
#    'tstscrapy.pipelines.TstscrapyPipeline': 300,
# }

# Enable and configure the AutoThrottle extension (disabled by default)
# See https://docs.scrapy.org/en/latest/topics/autothrottle.html
#AUTOTHROTTLE_ENABLED = True
# The initial download delay
#AUTOTHROTTLE_START_DELAY = 5
# The maximum download delay to be set in case of high latencies
#AUTOTHROTTLE_MAX_DELAY = 60
# The average number of requests Scrapy should be sending in parallel to
# each remote server
#AUTOTHROTTLE_TARGET_CONCURRENCY = 1.0
# Enable showing throttling stats for every response received:
#AUTOTHROTTLE_DEBUG = False

# Enable and configure HTTP caching (disabled by default)
# See https://docs.scrapy.org/en/latest/topics/downloader-middleware.html#httpcache-middleware-settings
#HTTPCACHE_ENABLED = True
#HTTPCACHE_EXPIRATION_SECS = 0
#HTTPCACHE_DIR = 'httpcache'
#HTTPCACHE_IGNORE_HTTP_CODES = []
#HTTPCACHE_STORAGE = 'scrapy.extensions.httpcache.FilesystemCacheStorage'

DOWNLOAD_TIMEOUT = 10
PYTHON

```

Contributing
=============
- Fork the repo
- Create a feature branch
- Push your changes
- Create a pull request

License
========
Apache License v2.0

Author
=======
- Mohamed Al Ashaal
- Software Engineer
- m7medalash3al@gmail.com
