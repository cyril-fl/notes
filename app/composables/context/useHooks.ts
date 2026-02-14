/* Hooks interface -- should be the same across divers projects */
/* Hooks builder */
export type HookRegistrar = () => void;
export type HookRegistrarFactory = () => HookRegistrar;

interface HooksBuilder {
  register(factory: HookRegistrarFactory): HooksBuilder;
  build(): void;
}

export function useHooksBuilder(): HooksBuilder {
  const factories: HookRegistrarFactory[] = [];

  const builder: HooksBuilder = {
    register(factory: HookRegistrarFactory) {
      factories.push(factory);
      return builder;
    },
    build() {
      factories.forEach((factory) => factory()());
    },
  };

  return builder;
}
