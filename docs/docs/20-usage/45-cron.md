# Cron

To enable cron jobs you need at least have push access to the repo.

:::warning
by default cron wont trigger any step in pipelines, as the default event branch filters it.  
Read more at: [pipeline-syntax#event](/docs/usage/pipeline-syntax#event)
:::

1. So to start add the event filter to all steps where you like to run:

    ```diff
     pipeline:
       sync_locales:
         image: weblate_sync
         settings:
           url: example.com
           token:
             from_secret: weblate_token
    +    when:
    +      event: cron
    ```

2. Create a new cron job at repo settings

    ![cron settings](./cron-settings.png)

    Schedule syntax can be found at https://pkg.go.dev/github.com/robfig/cron?utm_source=godoc#hdr-CRON_Expression_Format.  
    Examples: `@every 5m`, `@daily`, `0 30 * * * *` ...
