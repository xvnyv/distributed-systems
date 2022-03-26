import CAddItemModal from "./components/CAddItemModal";
import CTable from "./components/CTable";
import { Container, Grid, GridItem, Box, Input, FormLabel, Button } from "@chakra-ui/react";
import { useReducer, useState } from "react";
import { reducer } from "./reducers/ItemReducer";

var sampleItemsArray = {
  1: {
    id: 1,
    Name: "pencil",
    Quantity: 2,
  },
  3: {
    id: 3,
    Name: "paper",
    Quantity: 1,
  },
};

function App(itemArray) {
  var items =
    Object.keys(itemArray).length === 0 ? sampleItemsArray : itemArray;
  const [state, dispatch] = useReducer(reducer, items);

  const [user, setUser] = useState(0);

  
  return (
    <div>
      <Grid templateColumns="repeat(6, 1fr)" gap={2}>
        <GridItem colSpan={2} padding={3}>
          <Box borderRadius={10} border="1px" borderColor="blackAlpha.100" padding={5}>
            <CAddItemModal state={state} dispatch={dispatch}/>
            <FormLabel marginTop={10} marginLeft={2}>Change User</FormLabel>
            <Input placeholder="e.g, '5'" onChange={e => {setUser(e.target.value)}}/>
            <Button marginTop={3}>Find User</Button>
          </Box>
        </GridItem>
        <GridItem colSpan={4} padding={3}>
        <Box borderRadius={10} border="1px" borderColor="blackAlpha.100" padding={5}>

          <CTable state={state} dispatch={dispatch} />
          </Box>
        </GridItem>
      </Grid>
    </div>
  );
}

export default App;
