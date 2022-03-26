import {
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  TableCaption,
  Button,
  useDisclosure,
  AlertDialog,
  AlertDialogOverlay,
  AlertDialogFooter,
  AlertDialogBody,
  AlertDialogContent,
  AlertDialogHeader,
} from "@chakra-ui/react";
import { AddIcon, MinusIcon } from "@chakra-ui/icons";
//test
import { ITEM_ACTIONS } from "../reducers/ItemReducer";
import React, { useState } from "react";

const CTable = ({ state, dispatch }) => {
  //delete alert
  const { isOpen, onOpen, onClose } = useDisclosure();
  const cancelRef = React.useRef();
  const [toDelete, setToDelete] = useState();
  //return empty cart if no items
  if (Object.keys(state).length === 0) {
    return (
      <Table variant="simple">
        <TableCaption>Shopping Cart for User: {state.UserID} Empty</TableCaption>
      </Table>
    );
  }

  //itemIds and attributes init
  var itemIds = Object.keys(state.Item);
  var itemAttributes = Object.keys(state.Item[itemIds[0]]);

  //dispatching + , - , and del functions upon button press
  const itemAdd = (id) => {
    dispatch({ type: ITEM_ACTIONS.INCREMENT, payload: id });
  };
  const itemSubtract = (id) => {
    if (state.Item[id].Quantity === 1) {
      setToDelete(id);
      onOpen();
      return;
    }
    dispatch({ type: ITEM_ACTIONS.DECREMENT, payload: id });
  };
  const itemDelete = () => {
    dispatch({ type: ITEM_ACTIONS.DELETE, payload: toDelete });
    onClose();
  };

  return (
    <>
      <Table variant="simple">
        <TableCaption>Shopping Cart for User: {state.UserID}</TableCaption>
        <Thead>
          <Tr>
            {/* mapping item attributes into the column headers */}
            {itemAttributes.map((n) => (
              <Th>{n}</Th>
            ))}
            <Th align="centre">Change</Th>
          </Tr>
        </Thead>

        <Tbody>
          {/* Mapping items array into row */}
          {itemIds.map((n) => (
            <Tr>
              {itemAttributes.map((x) => (
                <Td>{state.Item[n][x]}</Td>
              ))}
              <Td>
                <Button marginRight={5} onClick={(e) => itemAdd(n)}>
                  <AddIcon />
                </Button>
                <Button onClick={(e) => itemSubtract(n)}>
                  <MinusIcon />
                </Button>
              </Td>
            </Tr>
          ))}
        </Tbody>
      </Table>

      {/* alert dialog for deleting item */}
      <AlertDialog isOpen={isOpen} leastDestructiveRef={cancelRef} onClose={onClose}>
        <AlertDialogOverlay>
          <AlertDialogContent>
            <AlertDialogHeader fontSize="lg" fontWeight="bold">
              Delete Item
            </AlertDialogHeader>

            <AlertDialogBody>Are you sure? You can't undo this action afterwards.</AlertDialogBody>

            <AlertDialogFooter>
              <Button ref={cancelRef} onClick={onClose}>
                Cancel
              </Button>
              <Button colorScheme="red" onClick={itemDelete} ml={3}>
                Delete
              </Button>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialogOverlay>
      </AlertDialog>
    </>
  );
};

export default CTable;
