import CAddItemModal from "./components/CAddItemModal";
import CTable from "./components/CTable";
import { Grid, GridItem, Box, Input, FormLabel, Button, useToast } from "@chakra-ui/react";
import React, { useReducer, useState } from "react";
import { reducer } from "./reducers/ItemReducer";
import GetDataAxios from "./axios/GetData";
var clientCartObject = {
  UserID: "SAMPLE",
  Item: {
    1: {
      Id: 1,
      Name: "pencil",
      Quantity: 2,
    },
    3: {
      Id: 3,
      Name: "paper",
      Quantity: 1,
    },
  },
  VectorClock: [1, 2, 3, 4],
};

function App() {
  const [state, dispatch] = useReducer(reducer, clientCartObject);
  const toast = useToast();
  const toastIdRef = React.useRef();

  const [userId, setUserId] = useState();

  const CheckEnter = (e) => {
    setUserId(e.target.value);

    console.log(e.target.value);
    if (e.key === "Enter") {
      console.log("hi");
      GetDataAxios(e.target.value, dispatch);
      return;
    }
    if (isNaN(e.key)) {
      toastIdRef.current = toast({
        title: "Input Error",
        status: "warning",
        description: "User's Id should be a number",
        duration: 2000,
        isClosable: true,
        position: "top-left",
      });
    }
  };

  return (
    <div>
      <Grid templateColumns="repeat(6, 1fr)" gap={2}>
        <GridItem colSpan={2} padding={3}>
          <Box borderRadius={10} border="1px" borderColor="blackAlpha.100" padding={5}>
            <FormLabel marginLeft={2}>Change User</FormLabel>
            <Input placeholder="e.g, '5'" onKeyPress={(e) => CheckEnter(e)} type="number" />
            <Button
              marginTop={3}
              marginBottom={10}
              width="100%"
              colorScheme={"pink"}
              variant="ghost"
              border="1px"
              borderColor="pink.100"
              onClick={() => {
                GetDataAxios(userId, dispatch);
              }}
            >
              Find User
            </Button>
            <CAddItemModal state={state} dispatch={dispatch} />
          </Box>
        </GridItem>
        <GridItem colSpan={4} padding={3}>
          <Box borderRadius={10} border="1px" borderColor="blackAlpha.100" padding={5}>
            {/* Table for to show shopping cart items */}
            <CTable state={state} dispatch={dispatch} />
          </Box>
        </GridItem>
      </Grid>
    </div>
  );
}

export default App;
