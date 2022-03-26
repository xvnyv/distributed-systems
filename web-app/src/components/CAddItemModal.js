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

import { PhoneIcon, AddIcon, WarningIcon } from "@chakra-ui/icons";

import React, { useState } from "react";
import { ACTION_TYPES } from "../reducers/ItemReducer";

const CAddItemModal = ({ state, dispatch }) => {
  const { isOpen, onOpen, onClose } = useDisclosure();

  const [itemId, setItemId] = useState();
  const [itemName, setItemName] = useState();

  const addItem = () => {
    dispatch({type: ACTION_TYPES.ADD_ITEM, payload: {itemId, itemName}})
    onClose()
  };

  return (
    <>
      <Button onClick={onOpen} width="100%">
        New Item
      </Button>

      <Modal isOpen={isOpen} onClose={onClose}>
        <ModalOverlay />
        <ModalContent>
          <ModalHeader>New Item</ModalHeader>
          <ModalCloseButton />
          <ModalBody>
            <FormControl>
              <FormLabel>Item ID</FormLabel>
              <Input
                onChange={(e) => {
                  setItemId(e.target.value);
                }}
                placeholder="e.g, 6"
              />
            </FormControl>

            <FormControl mt={4}>
              <FormLabel>Item Name</FormLabel>
              <Input
                onChange={(e) => {
                  setItemName(e.target.value);
                }}
                placeholder="e.g, banana"
                onSubmit={addItem}
              />
            </FormControl>
          </ModalBody>

          <ModalFooter>
            <Button colorScheme="blue" mr={3} onClick={addItem}>
              save
            </Button>
            <Button variant="ghost" onClick={onClose}>cancel</Button>
          </ModalFooter>
        </ModalContent>
      </Modal>
    </>
  );
};

export default CAddItemModal;
