# Forges

## Supported features

| Feature                                                          | [GitHub](20-github.md) | [Gitea](30-gitea.md) | [Forgejo](35-forgejo.md) | [Gitlab](40-gitlab.md) | [Bitbucket](50-bitbucket.md) | [Bitbucket Datacenter](60-bitbucket_datacenter.md) |
| ---------------------------------------------------------------- | :--------------------: | :------------------: | :----------------------: | :--------------------: | :--------------------------: | :------------------------------------------------: |
| Event: Push                                                      |   :white_check_mark:   |  :white_check_mark:  |    :white_check_mark:    |   :white_check_mark:   |      :white_check_mark:      |                 :white_check_mark:                 |
| Event: Tag                                                       |   :white_check_mark:   |  :white_check_mark:  |    :white_check_mark:    |   :white_check_mark:   |      :white_check_mark:      |                 :white_check_mark:                 |
| Event: Pull-Request                                              |   :white_check_mark:   |  :white_check_mark:  |    :white_check_mark:    |   :white_check_mark:   |      :white_check_mark:      |                 :white_check_mark:                 |
| Event: Release                                                   |   :white_check_mark:   |  :white_check_mark:  |    :white_check_mark:    |   :white_check_mark:   |             :x:              |                        :x:                         |
| Event: Deploy¹                                                   |   :white_check_mark:   |         :x:          |           :x:            |          :x:           |             :x:              |                        :x:                         |
| [Multiple workflows](../../../20-usage/25-workflows.md)          |   :white_check_mark:   |  :white_check_mark:  |    :white_check_mark:    |   :white_check_mark:   |      :white_check_mark:      |                 :white_check_mark:                 |
| [when.path filter](../../../20-usage/20-workflow-syntax.md#path) |   :white_check_mark:   |  :white_check_mark:  |    :white_check_mark:    |   :white_check_mark:   |             :x:              |                        :x:                         |

¹ The deployment event can be triggered for all forges from Woodpecker directly. However, only GitHub can trigger them using webhooks.
