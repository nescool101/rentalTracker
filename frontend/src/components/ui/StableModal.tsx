import { ReactNode, MouseEvent } from 'react';
import { Modal, ModalProps, Box } from '@mantine/core';

type StableModalProps = ModalProps & {
  children: ReactNode;
};

/**
 * A stable modal component that handles click events properly
 * to prevent the modal from disappearing when inputs are clicked
 */
export function StableModal({ children, overlayProps, onClose, ...props }: StableModalProps) {
  // Prevent the modal from closing unintentionally
  const handleOverlayClick = (e: MouseEvent) => {
    // Stop propagation for all clicks on the overlay or modal
    e.stopPropagation();
    
    // Only close if clicking directly on the overlay background, not on content
    if (e.currentTarget === e.target) {
      onClose();
    }
  };

  // Handle click inside modal content to prevent propagation
  const handleModalContentClick = (e: MouseEvent) => {
    // Stop event from bubbling up to prevent modal closing
    e.stopPropagation();
  };

  // Create a safer onClose handler to ensure it's called correctly
  const safeOnClose = () => {
    // Only call onClose if it exists
    if (typeof onClose === 'function') {
      onClose();
    }
  };

  // Combine custom overlay props with the ones passed in
  const combinedOverlayProps = {
    ...overlayProps,
    onClick: handleOverlayClick,
  };

  return (
    <Modal
      {...props}
      onClose={safeOnClose}
      trapFocus
      withCloseButton
      closeOnClickOutside={false}
      closeOnEscape={true}
      overlayProps={combinedOverlayProps}
      styles={{
        ...(props.styles as any),
        // Ensure clicks inside content don't bubble
        content: {
          ...((props.styles as any)?.content || {}),
          pointerEvents: 'auto',
        },
        // Ensure modal remains visible
        overlay: {
          ...((props.styles as any)?.overlay || {}),
          pointerEvents: 'auto',
        },
        // Prevent bubbling for any internal elements
        body: {
          ...((props.styles as any)?.body || {}),
          pointerEvents: 'auto',
        },
        // Ensure inner elements get clicks
        inner: {
          ...((props.styles as any)?.inner || {}),
          pointerEvents: 'auto',
        },
      }}
    >
      <Box p="md" onClick={handleModalContentClick} style={{ pointerEvents: 'auto' }}>
        {children}
      </Box>
    </Modal>
  );
} 