'use client';

import type { OryFlow, OryUiNode } from '@/types/ory';

function renderNode(node: OryUiNode) {
  const name = node.attributes.name;
  const type = node.attributes.type ?? 'text';
  const label = node.meta?.label?.text ?? name ?? 'Field';
  const messages = node.messages ?? [];
  const isOIDCButton = type === 'submit' && node.group === 'oidc';

  if (!name) {
    return null;
  }

  if (type === 'hidden' || type === 'submit') {
    if (type === 'submit') {
      return (
        <button
          className={isOIDCButton ? 'ghost-button social-button' : 'candy-button'}
          type="submit"
          name={name}
          value={String(node.attributes.value ?? '')}
          formNoValidate={isOIDCButton}
        >
          {label}
        </button>
      );
    }

    return <input type="hidden" name={name} value={String(node.attributes.value ?? '')} />;
  }

  return (
    <label className="field" key={`${name}-${type}`}>
      <span className="field__label">{label}</span>
      <input
        className="field__input"
        name={name}
        type={type}
        defaultValue={typeof node.attributes.value === 'string' ? node.attributes.value : undefined}
        required={node.attributes.required}
        disabled={node.attributes.disabled}
      />
      {messages.length > 0 && (
        <span className="field__message">{messages.map((message) => message.text).join(' ')}</span>
      )}
    </label>
  );
}

export function OryFlowForm({
  flow,
  title,
  subtitle,
}: {
  flow: OryFlow;
  title: string;
  subtitle: string;
}) {
  const generalMessages = flow.ui.messages ?? [];
  const hiddenNodes = flow.ui.nodes.filter((node) => node.attributes.type === 'hidden');
  const oidcNodes = flow.ui.nodes.filter((node) => node.group === 'oidc' && node.attributes.type === 'submit');
  const primaryNodes = flow.ui.nodes.filter(
    (node) => node.attributes.type !== 'hidden' && !(node.group === 'oidc' && node.attributes.type === 'submit'),
  );

  return (
    <div className="auth-card glass-card">
      <div className="section-heading">
        <span className="pill">Ory-powered access</span>
        <h1>{title}</h1>
        <p>{subtitle}</p>
      </div>

      {generalMessages.length > 0 && (
        <div className="message-stack">
          {generalMessages.map((message) => (
            <div className="toast toast--warning" key={message.id}>
              {message.text}
            </div>
          ))}
        </div>
      )}

      <form action={flow.ui.action} method={flow.ui.method} className="auth-form">
        {hiddenNodes.map((node) => (
          <div key={`${node.attributes.name ?? 'node'}-${node.type}-${node.group ?? 'default'}`}>
            {renderNode(node)}
          </div>
        ))}

        {primaryNodes.map((node) => (
          <div key={`${node.attributes.name ?? 'node'}-${node.type}-${node.group ?? 'default'}`}>
            {renderNode(node)}
          </div>
        ))}

        {oidcNodes.length > 0 && (
          <div className="social-auth-block">
            <div className="social-auth-divider">
              <span>or continue with</span>
            </div>
            <div className="social-auth-grid">
              {oidcNodes.map((node) => (
                <div key={`${node.attributes.name ?? 'node'}-${node.type}-${node.group ?? 'default'}`}>
                  {renderNode(node)}
                </div>
              ))}
            </div>
          </div>
        )}
      </form>
    </div>
  );
}
