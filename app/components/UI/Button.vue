<script lang="ts" setup>
import { tv, type VariantProps } from 'tailwind-variants';
import { Primitive } from 'reka-ui';
import { NuxtLink } from '#components';

const theme = tv({
  base: 'inline-flex items-center justify-center font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 cursor-pointer',
  variants: {
    size: {
      xs: 'text-xs h-6 px-2 gap-1 rounded-sm',
      sm: 'text-xs h-7 px-2.5 gap-1.5 rounded-sm',
      md: 'text-sm h-8 px-3 gap-1.5 rounded-md',
      lg: 'text-sm h-9 px-3.5 gap-2 rounded-md',
      xl: 'text-base h-10 px-4 gap-2 rounded-lg',
      '2xl': 'text-base h-11 px-5 gap-2.5 rounded-lg',
      '3xl': 'text-lg h-12 px-6 gap-3 rounded-xl',
    },
    color: {
      primary: '',
      secondary: '',
      neutral: '',
      success: '',
      info: '',
      warning: '',
      error: '',
    },
    variant: {
      solid: '',
      subtle: '',
      soft: '',
      ghost: '',
      outline: '',
      link: 'underline-offset-4 hover:underline',
    },
    square: {
      true: 'aspect-square p-0',
    },
  },
  compoundVariants: [
    // PRIMARY
    {
      color: 'primary',
      variant: 'solid',
      class:
        'bg-primary-600 text-white hover:bg-primary-700 focus-visible:ring-primary-500',
    },
    {
      color: 'primary',
      variant: 'subtle',
      class:
        'bg-primary-100 text-primary-700 hover:bg-primary-200 dark:bg-primary-900 dark:text-primary-300 dark:hover:bg-primary-800',
    },
    {
      color: 'primary',
      variant: 'soft',
      class:
        'bg-primary-50 text-primary-600 hover:bg-primary-100 dark:bg-primary-950 dark:text-primary-400 dark:hover:bg-primary-900',
    },
    {
      color: 'primary',
      variant: 'ghost',
      class:
        'text-primary-600 hover:bg-primary-100 dark:text-primary-400 dark:hover:bg-primary-900',
    },
    {
      color: 'primary',
      variant: 'outline',
      class:
        'border border-primary-500 text-primary-600 hover:bg-primary-50 dark:text-primary-400 dark:hover:bg-primary-950',
    },
    {
      color: 'primary',
      variant: 'link',
      class: 'text-primary-600 dark:text-primary-400',
    },

    // SECONDARY
    {
      color: 'secondary',
      variant: 'solid',
      class:
        'bg-secondary-600 text-white hover:bg-secondary-700 focus-visible:ring-secondary-500',
    },
    {
      color: 'secondary',
      variant: 'subtle',
      class:
        'bg-secondary-100 text-secondary-700 hover:bg-secondary-200 dark:bg-secondary-800 dark:text-secondary-300 dark:hover:bg-secondary-700',
    },
    {
      color: 'secondary',
      variant: 'soft',
      class:
        'bg-secondary-50 text-secondary-600 hover:bg-secondary-100 dark:bg-secondary-900 dark:text-secondary-400 dark:hover:bg-secondary-800',
    },
    {
      color: 'secondary',
      variant: 'ghost',
      class:
        'text-secondary-600 hover:bg-secondary-100 dark:text-secondary-400 dark:hover:bg-secondary-800',
    },
    {
      color: 'secondary',
      variant: 'outline',
      class:
        'border border-secondary-400 text-secondary-600 hover:bg-secondary-50 dark:border-secondary-600 dark:text-secondary-400 dark:hover:bg-secondary-900',
    },
    {
      color: 'secondary',
      variant: 'link',
      class: 'text-secondary-600 dark:text-secondary-400',
    },

    // NEUTRAL
    {
      color: 'neutral',
      variant: 'solid',
      class:
        'bg-neutral-800 text-white hover:bg-neutral-900 dark:bg-neutral-200 dark:text-neutral-900 dark:hover:bg-neutral-100 focus-visible:ring-neutral-500',
    },
    {
      color: 'neutral',
      variant: 'subtle',
      class:
        'bg-neutral-100 text-neutral-700 hover:bg-neutral-200 dark:bg-neutral-800 dark:text-neutral-300 dark:hover:bg-neutral-700',
    },
    {
      color: 'neutral',
      variant: 'soft',
      class:
        'bg-neutral-50 text-neutral-600 hover:bg-neutral-100 dark:bg-neutral-900 dark:text-neutral-400 dark:hover:bg-neutral-800',
    },
    {
      color: 'neutral',
      variant: 'ghost',
      class:
        'text-neutral-600 hover:bg-neutral-100 dark:text-neutral-400 dark:hover:bg-neutral-800',
    },
    {
      color: 'neutral',
      variant: 'outline',
      class:
        'border border-neutral-300 text-neutral-700 hover:bg-neutral-50 dark:border-neutral-600 dark:text-neutral-300 dark:hover:bg-neutral-800',
    },
    {
      color: 'neutral',
      variant: 'link',
      class: 'text-neutral-600 dark:text-neutral-400',
    },

    // SUCCESS
    {
      color: 'success',
      variant: 'solid',
      class:
        'bg-success-600 text-white hover:bg-success-700 focus-visible:ring-success-500',
    },
    {
      color: 'success',
      variant: 'subtle',
      class:
        'bg-success-100 text-success-700 hover:bg-success-200 dark:bg-success-900 dark:text-success-300 dark:hover:bg-success-800',
    },
    {
      color: 'success',
      variant: 'soft',
      class:
        'bg-success-50 text-success-600 hover:bg-success-100 dark:bg-success-950 dark:text-success-400 dark:hover:bg-success-900',
    },
    {
      color: 'success',
      variant: 'ghost',
      class:
        'text-success-600 hover:bg-success-100 dark:text-success-400 dark:hover:bg-success-900',
    },
    {
      color: 'success',
      variant: 'outline',
      class:
        'border border-success-500 text-success-600 hover:bg-success-50 dark:text-success-400 dark:hover:bg-success-950',
    },
    {
      color: 'success',
      variant: 'link',
      class: 'text-success-600 dark:text-success-400',
    },

    // INFO
    {
      color: 'info',
      variant: 'solid',
      class:
        'bg-info-600 text-white hover:bg-info-700 focus-visible:ring-info-500',
    },
    {
      color: 'info',
      variant: 'subtle',
      class:
        'bg-info-100 text-info-700 hover:bg-info-200 dark:bg-info-900 dark:text-info-300 dark:hover:bg-info-800',
    },
    {
      color: 'info',
      variant: 'soft',
      class:
        'bg-info-50 text-info-600 hover:bg-info-100 dark:bg-info-950 dark:text-info-400 dark:hover:bg-info-900',
    },
    {
      color: 'info',
      variant: 'ghost',
      class:
        'text-info-600 hover:bg-info-100 dark:text-info-400 dark:hover:bg-info-900',
    },
    {
      color: 'info',
      variant: 'outline',
      class:
        'border border-info-500 text-info-600 hover:bg-info-50 dark:text-info-400 dark:hover:bg-info-950',
    },
    {
      color: 'info',
      variant: 'link',
      class: 'text-info-600 dark:text-info-400',
    },

    // WARNING
    {
      color: 'warning',
      variant: 'solid',
      class:
        'bg-warning-500 text-white hover:bg-warning-600 focus-visible:ring-warning-500',
    },
    {
      color: 'warning',
      variant: 'subtle',
      class:
        'bg-warning-100 text-warning-700 hover:bg-warning-200 dark:bg-warning-900 dark:text-warning-300 dark:hover:bg-warning-800',
    },
    {
      color: 'warning',
      variant: 'soft',
      class:
        'bg-warning-50 text-warning-600 hover:bg-warning-100 dark:bg-warning-950 dark:text-warning-400 dark:hover:bg-warning-900',
    },
    {
      color: 'warning',
      variant: 'ghost',
      class:
        'text-warning-600 hover:bg-warning-100 dark:text-warning-400 dark:hover:bg-warning-900',
    },
    {
      color: 'warning',
      variant: 'outline',
      class:
        'border border-warning-500 text-warning-600 hover:bg-warning-50 dark:text-warning-400 dark:hover:bg-warning-950',
    },
    {
      color: 'warning',
      variant: 'link',
      class: 'text-warning-600 dark:text-warning-400',
    },

    // ERROR
    {
      color: 'error',
      variant: 'solid',
      class:
        'bg-error-600 text-white hover:bg-error-700 focus-visible:ring-error-500',
    },
    {
      color: 'error',
      variant: 'subtle',
      class:
        'bg-error-100 text-error-700 hover:bg-error-200 dark:bg-error-900 dark:text-error-300 dark:hover:bg-error-800',
    },
    {
      color: 'error',
      variant: 'soft',
      class:
        'bg-error-50 text-error-600 hover:bg-error-100 dark:bg-error-950 dark:text-error-400 dark:hover:bg-error-900',
    },
    {
      color: 'error',
      variant: 'ghost',
      class:
        'text-error-600 hover:bg-error-100 dark:text-error-400 dark:hover:bg-error-900',
    },
    {
      color: 'error',
      variant: 'outline',
      class:
        'border border-error-500 text-error-600 hover:bg-error-50 dark:text-error-400 dark:hover:bg-error-950',
    },
    {
      color: 'error',
      variant: 'link',
      class: 'text-error-600 dark:text-error-400',
    },

    // Square sizes (override padding)
    { square: true, size: 'xs', class: 'w-6' },
    { square: true, size: 'sm', class: 'w-7' },
    { square: true, size: 'md', class: 'w-8' },
    { square: true, size: 'lg', class: 'w-9' },
    { square: true, size: 'xl', class: 'w-10' },
    { square: true, size: '2xl', class: 'w-11' },
    { square: true, size: '3xl', class: 'w-12' },
  ],
  defaultVariants: {
    size: 'md',
    color: 'primary',
    variant: 'solid',
  },
});

type ButtonVariants = VariantProps<typeof theme>;

interface ButtonProps {
  type: 'button' | 'submit' | 'reset';
  to: string;
  label: string;
  icon: string;
  trailing: boolean;
  leading: boolean;
  disabled: boolean;
  square: boolean;
  size: ButtonVariants['size'];
  color: ButtonVariants['color'];
  variant: ButtonVariants['variant'];
  class: CssClass;
}

const props = withDefaults(defineProps<Partial<ButtonProps>>(), {
  type: 'button',
  to: undefined,
  label: '',
  icon: '',
  trailing: false,
  leading: true,
  square: false,
  disabled: false,
  size: 'md',
  color: 'primary',
  variant: 'solid',
  class: '',
});

const iconSize = computed(() => {
  const sizes: Record<string, string> = {
    xs: '12',
    sm: '14',
    md: '16',
    lg: '18',
    xl: '20',
    '2xl': '22',
    '3xl': '24',
  };
  return sizes[props.size ?? 'md'];
});

const className = computed(() =>
  theme({
    size: props.size,
    color: props.color,
    variant: props.variant,
    square: props.square,
    class: props.class as string,
  })
);

const buttonProps = computed(() => {
  return props.to
    ? {
        as: NuxtLink,
        to: props.to,
      }
    : {
        as: 'button',
        type: props.type,
        disabled: props.disabled,
      };
});
</script>

<template>
  <Primitive :class="className" v-bind="buttonProps">
    <Icon v-if="icon && leading && !trailing" :name="icon" :size="iconSize" />
    <slot v-if="!square">{{ label }}</slot>
    <Icon
      v-if="icon && ((trailing && !square) || square)"
      :name="icon"
      :size="iconSize"
    />
  </Primitive>
</template>
