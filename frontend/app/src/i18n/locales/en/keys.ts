const keys = {
  eyebrow: "API Keys",
  title: "Manage access keys",
  description: "Create, disable, or delete your platform API keys. A new key is shown in full only once when created.",
  createTitle: "Create a new key",
  namePlaceholder: "Key name, for example: Production",
  defaultName: "Default Key",
  createAction: "Create",
  createSuccess: "API key created",
  copyBaseURLSuccess: "base_url copied",
  updateSuccess: "API key updated",
  deleteSuccess: "API key deleted",
  deleteConfirm: "Delete this API key? Requests using it will no longer be able to access the gateway.",
  emptyTitle: "No API keys yet",
  emptyDescription: "Create a key and point your OpenAI SDK base_url to the LingShu gateway.",
  emptyAction: "Create key",
  table: {
    name: "Name",
    mask: "Key",
    endpoints: "Allowed Endpoints",
    status: "Status",
    createdAt: "Created At",
    actions: "Actions",
    copyBaseURL: "Copy base_url",
    docs: "Examples",
    edit: "Edit",
    delete: "Delete"
  },
  endpoints: {
    title: "Endpoint permissions",
    all: "Allow all endpoints",
    hint: "Leave empty to allow every gateway endpoint, or select only the interfaces this key needs.",
    chat: "OpenAI Chat",
    messages: "Anthropic Messages",
    embeddings: "Embeddings",
    models: "Model list"
  },
  edit: {
    title: "Edit key",
    save: "Save",
    cancel: "Cancel"
  },
  dialog: {
    title: "Copy the new key now",
    description: "The full key is shown only once and cannot be viewed again after closing.",
    saved: "I saved it",
    copied: "Key copied",
    ariaCopy: "Copy key"
  }
};

export default keys;
