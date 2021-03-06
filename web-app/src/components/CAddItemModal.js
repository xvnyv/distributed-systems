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
  useToast,
} from "@chakra-ui/react";

import React, { useState } from "react";
import { SendPostRequest } from "../http_helpers/PostGetRequesters";
import { CLIENTCART_ACTIONS } from "../reducers/ClientCartReducer";

const CAddItemModal = ({
  state,
  dispatch,
  appToast,
  appToastRef,
  clientId,
}) => {
  // to control state of modal
  const { isOpen, onOpen, onClose } = useDisclosure();

  // to remember the state of itemId and itemName to be created
  const [itemId, setItemId] = useState();
  const [itemName, setItemName] = useState();

  //toast to provide feedback for wrong input or item already existing
  const toast = useToast();
  const toastIdRef = React.useRef();

  // dispatching new item call if all validation passes
  const addItem = () => {
    if (itemId === "" || itemName === "") {
      toastIdRef.current = toast({
        title: "Input Error",
        status: "error",
        description: "Please fill in all fields",
        duration: 2000,
        isClosable: true,
        position: "top-left",
      });
      return;
    }
    if (typeof state.Item[itemId] !== "undefined") {
      console.log(state.Item[itemId]);
      toastIdRef.current = toast({
        title: "Input Error",
        status: "error",
        description: "Item ID already exists",
        duration: 2000,
        isClosable: true,
        position: "top-left",
      });
      return;
    }
    var newState = {
      ...state,
      Item: {
        ...state.Item,
        [parseInt(itemId)]: {
          Id: parseInt(itemId),
          Name: itemName,
          Quantity: 1,
        },
      },
    };
    dispatch({ type: CLIENTCART_ACTIONS.UPDATE_STATE, payload: newState });
    SendPostRequest(newState, appToast, appToastRef, dispatch, clientId);
    onClose();
  };

  // add item if keypress == enter
  const CheckEnter = (e) => {
    if (e.key === "Enter") {
      addItem();
    }
  };
  const CheckNumber = (e) => {
    if (isNaN(e.key)) {
      toastIdRef.current = toast({
        title: "Input Error",
        status: "warning",
        description: "Item ID should be a number",
        duration: 2000,
        isClosable: true,
        position: "top-left",
      });
    }
  };

  return (
    <>
      <Button
        onClick={onOpen}
        width="100%"
        colorScheme={"teal"}
        variant="ghost"
        border="1px"
        borderColor="teal.100"
      >
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
                onKeyPress={(e) => {
                  CheckNumber(e);
                  CheckEnter(e);
                }}
                placeholder="e.g, 6"
                type="number"
              />
            </FormControl>

            <FormControl mt={4}>
              <FormLabel>Item Name</FormLabel>
              <Input
                onChange={(e) => {
                  setItemName(e.target.value);
                }}
                onKeyPress={(e) => CheckEnter(e)}
                placeholder="e.g, banana"
                onSubmit={addItem}
              />
            </FormControl>
          </ModalBody>

          <ModalFooter>
            <Button colorScheme="blue" mr={3} onClick={addItem}>
              save
            </Button>
            <Button variant="ghost" onClick={onClose}>
              cancel
            </Button>
          </ModalFooter>
        </ModalContent>
      </Modal>
    </>
  );
};

export default CAddItemModal;
