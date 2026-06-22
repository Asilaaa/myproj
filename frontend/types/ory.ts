export type OryUiNode = {
  attributes: {
    name?: string;
    type?: string;
    value?: string | number | readonly string[];
    required?: boolean;
    disabled?: boolean;
  };
  messages?: Array<{ id: number; text: string; type: string }>;
  meta?: {
    label?: {
      text?: string;
    };
  };
  group?: string;
  type: string;
};

export type OryFlow = {
  id: string;
  type: string;
  ui: {
    action: string;
    method: string;
    messages?: Array<{ id: number; text: string; type: string }>;
    nodes: OryUiNode[];
  };
};

export type OrySession = {
  active: boolean;
  identity: {
    id: string;
    traits?: Record<string, unknown> & {
      email?: string;
      name?: {
        first?: string;
        last?: string;
      };
    };
  };
};
