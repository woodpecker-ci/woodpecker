export type WoodpeckerPluginHeader = {
  name?: string;
  description?: string;
  icon?: string;
  url?: string;
};

export type WoodpeckerPluginIndexEntry = {
  name: string; // name of the plugin
  docs: string; // http url to the docs.md file
  verified?: boolean; // plugins maintained by trusted parties
};

export type WoodpeckerPlugin = WoodpeckerPluginHeader & {
  name: string;
  docs: string; // body of the docs .md file
  verified: boolean; // plugins maintained by trusted parties
};

export type Content = {
  plugins: WoodpeckerPlugin[];
};
