export type WoodpeckerPluginHeader = {
  name?: string; // name of the plugin
  description?: string; // short description of the plugin
  url?: string; // url of the plugin normally link to forge
  tags?: string[]; // tags to categorize the plugin
  author?: string; // author of the plugin
  icon?: string; // url pointing to an icon
  containerImage?: string; // name of a container image
  containerImageUrl?: string; // url to a container image registry
};

export type WoodpeckerPluginIndexEntry = {
  name: string; // name of the plugin
  docs: string; // http url to the docs.md file
  verified?: boolean; // plugins maintained by trusted parties
};

export type WoodpeckerPlugin = WoodpeckerPluginHeader & {
  name: string;
  docs: string; // body of the docs .md file
  verified: boolean; // we set verified to false when not explicitly set
};

export type Content = {
  plugins: WoodpeckerPlugin[];
};
