<template>
  <div>
    <el-input placeholder="请输入内容" v-model="content" class="input-with-select"
      style="margin-top: 15px;width:50%;margin-left: 15px;">
      <el-select v-model="select" slot="prepend" placeholder="请选择">
        <el-option label="单位" value="1"></el-option>
        <el-option label="价格" value="2"></el-option>
      </el-select>
      <el-button slot="append" icon="el-icon-search" @click="handleSearch()"></el-button>
    </el-input>

    <el-table :data="tableData" border style="margin-top: 15px;width:90%;margin-left: 15px;">
      <el-table-column label="商品ID" width="" align="center">
        <template slot-scope="scope">
          {{ scope.row.id }}
        </template>
      </el-table-column>
      <el-table-column label="拥有者" width="" align="center">
        <template slot-scope="scope">
          {{ scope.row.owner }}
          <el-popover />
        </template>
      </el-table-column>
      <el-table-column label="价格" width="" align="center">
        <template slot-scope="scope">
          {{ scope.row.price }}
          <el-popover />
        </template>
      </el-table-column>
      <el-table-column label="单位" width="" align="center">
        <template slot-scope="scope">
          {{ scope.row.amount }}
          <el-popover />
        </template>
      </el-table-column>
      <el-table-column label="操作" align="center">
        <template slot-scope="scope">
          <!-- 使用v-if或v-show来控制按钮的显示 -->
          <el-button v-if="scope.row.status === 1" size="mini" type="primary"
            @click="handleEdit(scope.$index, scope.row)">
            购买
          </el-button>
          <span v-else-if="scope.row.status===0">
            未上架
          </span>
          <span v-else-if="scope.row.status===2">
            锁定中
          </span>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog title="编辑订单信息" :visible.sync="dialogFormVisible">
      <el-form :model="form">
        <el-form-item label="拥有者">
          {{ form.owner }}
        </el-form-item>
        <el-form-item label="定价">
          {{ form.price }}
        </el-form-item>
        <el-form-item label="单位">
          {{ form.amount }}
        </el-form-item>
        <el-form-item label="购买者" :label-width="formLabelWidth">
          <el-input v-model="form.buyer" autocomplete="off"></el-input>
        </el-form-item>
        <el-form-item label="出价" :label-width="formLabelWidth">
          <el-input v-model="form.offer" autocomplete="off"></el-input>
        </el-form-item>
      </el-form>
      <div slot="footer" class="dialog-footer">
        <el-button @click="dialogFormVisible = false">取 消</el-button>
        <el-button type="primary" @click="handleSubmitProposal()">确 定</el-button>
      </div>
    </el-dialog>
  </div>
</template>


<style>
.el-select .el-input {
  width: 130px;
}

.input-with-select .el-input-group__prepend {
  background-color: #fff;
}
</style>
<script type="text/javascript">
import { listElec } from '@/api/elec';
import { addOrder } from '@/api/order';
export default {
  data() {
    return {
      dialogFormVisible: false,
      tableData: [],
      content: '',
      select: '',
      formLabelWidth: '',
      form: {
        index: '',
        id: '',
        buyer: '',
        owner: '',
        price: '',//定价
        amount: '',
        offer: ''//出价
      }
    }
  },
  methods: {
    loadData() {
      var response = listElec().then(
        response => {
          this.tableData = JSON.parse(response.data)
        }
      )
    },
    handleEdit(index, row) {
      var id = row.id;
      var owner = row.owner;
      // var buyer = row.buyer;
      var price = row.price;
      var amount = row.amount;
      // var offer = row.offer;
      this.dialogFormVisible = true;
      this.form.index = index;
      this.form.id = id;
      this.form.owner = owner;
      this.form.price = price;
      // this.from.offer=offer;
      this.form.amount = amount
    },
    handleSubmitProposal() {
      var id = this.form.id;
      var buyer = this.form.buyer;
      var seller = this.form.owner;
      var price = this.form.offer;
      if (id === null || id === ""
        || buyer === null || buyer === ""
        || seller === null || seller === ""
        || price === null || price === ""
      ) {
        this.$message({
          message: '请填写完成的信息',
          type: 'warning'
        });
      } else {
        //提交修改请求
        var data = {
          buyer: buyer,
          seller: seller,
          price: price,
          goodId: id
        }
        addOrder(data).then(
          response => {
            // this.tableData = JSON.parse(response.data)
            this.dialogFormVisible = false;
            this.$message({
              message: response.msg,
              type: 'success'
            });
          }
        );
      }
    }
  },
  mounted: function () {
    this.loadData()
  }
}
</script>
