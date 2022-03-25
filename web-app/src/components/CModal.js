import {
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalFooter,
  ModalBody,
  ModalCloseButton,
  Button,
  useDisclosure,
  FormControl,
  FormLabel,
  Input,

} from "@chakra-ui/react";

import { PhoneIcon, AddIcon, WarningIcon } from '@chakra-ui/icons'



import React from "react";

const CModal = ({state, dispatch}) => {
  const { isOpen, onOpen, onClose } = useDisclosure();
  return (
    <>
      <Button onClick={onOpen}>New Item</Button>

      <Modal isOpen={isOpen} onClose={onClose}>
        <ModalOverlay />
        <ModalContent>
          <ModalHeader>New Item</ModalHeader>
          <ModalCloseButton />
          <ModalBody>
          <FormControl>
              <FormLabel>Item ID</FormLabel>
              <Input  placeholder='e.g, 6' />
            </FormControl>

            <FormControl mt={4}>
              <FormLabel>Item Name</FormLabel>
              <Input placeholder='e.g, banana' />
            </FormControl>
          </ModalBody>

          <ModalFooter>
            <Button colorScheme="blue" mr={3} onClick={onClose}>
              save
            </Button>
            <Button variant="ghost">cancel</Button>
          </ModalFooter>
        </ModalContent>
      </Modal>
    </>
  );
};

export default CModal;